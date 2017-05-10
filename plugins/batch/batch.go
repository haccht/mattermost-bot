package batch

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"mattermost-bot"
)

type Plugin struct {
	bot      *mmbot.BotKit
	username string
	icon_url string
}

func NewPlugin(bot *mmbot.BotKit) *Plugin {
	p := &Plugin{bot: bot, username: "Batch scheduler"}
	p.refreshBatchTasks()
	return p
}

func (p *Plugin) HandleMessage(text, channel, username string) error {
	var re *regexp.Regexp

	// add batch task
	re = regexp.MustCompile(`(?i)^batch\s+add\s+(` + "`" + `.+` + "`" + `.+)$`)
	if re.MatchString(text) {
		submatch := re.FindSubmatch([]byte(text))
		p.addBatchTask(channel, string(submatch[1]))
		return nil
	}

	// delete batch task
	re = regexp.MustCompile(`(?i)^batch\s+del\s+(.*)$`)
	if re.MatchString(text) {
		submatch := re.FindSubmatch([]byte(text))
		p.delBatchTask(channel, string(submatch[1]))
		return nil
	}

	// list batch tasks
	re = regexp.MustCompile(`(?i)^batch\s+list$`)
	if re.MatchString(text) {
		p.listBatchTasks(channel)
		return nil
	}

	return nil
}

func (p *Plugin) Usage() string {
	usages := []string{
		`batch add ` + "`" + `<spec>` + "`" + ` <task>: Add a batch task.`,
		`batch del <id>: Delete the batch task.`,
		`batch list: List all batch tasks.`,
	}
	return strings.Join(usages, "\n")
}

func (p *Plugin) addBatchTask(channel, batchTask string) {
	// generate uniq batchId
	var batchId, batchKey string
	for {
		rand.Seed(time.Now().UnixNano())
		batchId = fmt.Sprint(rand.Intn(1000))
		batchKey = fmt.Sprintf("%s:%s", channel, batchId)
		if _, err := p.bot.Memory.Get(p, batchKey); err != nil {
			break
		}
	}

	if err := p.bot.Memory.Put(p, batchKey, batchTask); err != nil {
		message := fmt.Sprintf("Invalid batch '%s'\n%s", batchId, err.Error())
		p.bot.SendMessage(message, channel, p.username, p.icon_url)
	} else {
		if err := p.refreshBatchTasks(); err != nil {
			p.delBatchTask(channel, batchId)
			message := fmt.Sprintf("Failed to restart scheduler '%s'\n%s", batchTask, err.Error())
			p.bot.SendMessage(message, channel, p.username, p.icon_url)
		} else {
			message := "Added batch.\n"
			message += "```\n"
			message += fmt.Sprintf("%s: %s\n", batchId, batchTask)
			message += "```"
			p.bot.SendMessage(message, channel, p.username, p.icon_url)
		}
	}
}

func (p *Plugin) delBatchTask(channel, batchId string) {
	batchKey := fmt.Sprintf("%s:%s", channel, batchId)
	if batchTask, err := p.bot.Memory.Del(p, batchKey); err != nil {
		message := fmt.Sprintf("Invalid batch '%s'\n%s", batchId, err.Error())
		p.bot.SendMessage(message, channel, p.username, p.icon_url)
	} else {
		if err := p.refreshBatchTasks(); err != nil {
			message := fmt.Sprintf("Failed to restart scheduler '%s'\n%s", batchTask, err.Error())
			p.bot.SendMessage(message, channel, p.username, p.icon_url)
		} else {
			message := "Deleted batch.\n"
			message += "```\n"
			message += fmt.Sprintf("%s: %s\n", batchId, batchTask)
			message += "```"
			p.bot.SendMessage(message, channel, p.username, p.icon_url)
		}
	}
}

func (p *Plugin) listBatchTasks(channel string) {
	re := regexp.MustCompile(`^` + "`" + `\s*([^` + "`" + `]+)\s*` + "`" + `\s+(.+)$`)
	batchList := map[string]string{}

	// get tasks in the specified channel
	if list, err := p.bot.Memory.List(p); err == nil {
		for batchKey, batchTask := range list {
			submatch := re.FindSubmatch([]byte(batchTask))

			t1 := time.Now()
			t2, _ := p.parseTimeSpec(string(submatch[1]))

			if t2.After(t1) {
				if strings.HasPrefix(batchKey, channel+":") {
					batchList[batchKey] = batchTask
				}
			} else {
				// clean the outdated task
				p.bot.Memory.Del(p, batchKey)
			}
		}
	}

	if len(batchList) == 0 {
		message := "Could not find batchs."
		p.bot.SendMessage(message, channel, p.username, p.icon_url)
	} else {
		message := "```\n"
		for batchKey, batchTask := range batchList {
			batchId := batchKey[len(channel)+1:]
			message += fmt.Sprintf("%s: %s\n", batchId, batchTask)
		}
		message += "```"
		p.bot.SendMessage(message, channel, p.username, p.icon_url)
	}
}

func (p *Plugin) refreshBatchTasks() error {
	// get all tasks
	re := regexp.MustCompile(`^` + "`" + `\s*([^` + "`" + `]+)\s*` + "`" + `\s+(.+)$`)
	batchList := map[string]string{}
	batchList, _ = p.bot.Memory.List(p)

	for batchKey, batchTask := range batchList {
		submatch := re.FindSubmatch([]byte(batchTask))

		t1 := time.Now()
		t2, err := p.parseTimeSpec(string(submatch[1]))

		if err != nil {
			return err
		}

		if t2.After(t1) {
			go func() {
				channel := batchKey[:strings.Index(batchKey, ":")]

				select {
				case <-time.After(t2.Sub(t1)):
					p.bot.SendMessage(string(submatch[2]), channel, p.username, p.icon_url)
				}
			}()
		}
	}

	return nil
}

func (p *Plugin) parseTimeSpec(spec string) (time.Time, error) {
	now := time.Now()
	loc, _ := time.LoadLocation("Asia/Tokyo")

	var err error
	var parsedTime time.Time

	var re *regexp.Regexp

	re = regexp.MustCompile(`^\d\d\d\d/\d\d/\d\d\s+\d\d:\d\d$`)
	if re.MatchString(spec) {
		parsedTime, err = time.ParseInLocation("2006/01/02 15:04", spec, loc)
	}

	re = regexp.MustCompile(`^\d\d/\d\d\s+\d\d:\d\d$`)
	if re.MatchString(spec) {
		spec = fmt.Sprintf("%04d/%s", now.Year(), spec)
		parsedTime, err = time.ParseInLocation("2006/01/02 15:04", spec, loc)
	}

	re = regexp.MustCompile(`^\d\d:\d\d$`)
	if re.MatchString(spec) {
		spec = fmt.Sprintf("%04d/%02d/%02d %s", now.Year(), int(now.Month()), now.Day(), spec)
		parsedTime, err = time.ParseInLocation("2006/01/02 15:04", spec, loc)
	}

	if err != nil {
		return time.Time{}, fmt.Errorf("Could not parse datetime: %s", spec)
	} else {
		return parsedTime, nil
	}
}
