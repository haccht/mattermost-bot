package cron

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/robfig/cron"
	"mattermost-bot"
)

type Plugin struct {
	bot      *mmbot.BotKit
	cron     *cron.Cron
	username string
	icon_url string
}

func NewPlugin(bot *mmbot.BotKit) *Plugin {
	p := &Plugin{bot: bot, username: "Cron"}
	p.restartCronTasks()
	return p
}

func (p *Plugin) HandleMessage(text, channel, username string) error {
	var re *regexp.Regexp

	// add cron task
	re = regexp.MustCompile(`(?i)^cron\s+add\s+(` + "`" + `.+` + "`" + `.+)$`)
	if re.MatchString(text) {
		submatch := re.FindSubmatch([]byte(text))
		p.addCronTask(channel, string(submatch[1]))
		return nil
	}

	// delete cron task
	re = regexp.MustCompile(`(?i)^cron\s+del\s+(.*)$`)
	if re.MatchString(text) {
		submatch := re.FindSubmatch([]byte(text))
		p.delCronTask(channel, string(submatch[1]))
		return nil
	}

	// list cron tasks
	re = regexp.MustCompile(`(?i)^cron\s+list$`)
	if re.MatchString(text) {
		p.listCronTasks(channel)
		return nil
	}

	return nil
}

func (p *Plugin) Usage() string {
	usages := []string{
		`cron add ` + "`" + `<spec>` + "`" + ` <task>: Add a cron task.`,
		`cron del <id>: Delete the cron task.`,
		`cron list: List all cron tasks.`,
	}
	return strings.Join(usages, "\n")
}

func (p *Plugin) addCronTask(channel, cronTask string) {
	// generate uniq cronId and cronKey
	var cronId, cronKey string
	for {
		rand.Seed(time.Now().UnixNano())
		cronId = fmt.Sprint(rand.Intn(1000))
		cronKey = fmt.Sprintf("%s:%s", channel, cronId)
		if _, err := p.bot.Memory.Get(p, cronKey); err != nil {
			break
		}
	}

	if err := p.bot.Memory.Put(p, cronKey, cronTask); err != nil {
		message := fmt.Sprintf("Invalid cron task '%s'\n%s", cronId, err.Error())
		p.bot.SendMessage(message, channel, p.username, p.icon_url)
	} else {
		if err := p.restartCronTasks(); err != nil {
			p.delCronTask(channel, cronId)
			message := fmt.Sprintf("Failed to restart cron '%s'\n%s", cronTask, err.Error())
			p.bot.SendMessage(message, channel, p.username, p.icon_url)
		} else {
			message := "Added cron task.\n"
			message += "```\n"
			message += fmt.Sprintf("%s: %s\n", cronId, cronTask)
			message += "```"
			p.bot.SendMessage(message, channel, p.username, p.icon_url)
		}
	}
}

func (p *Plugin) delCronTask(channel, cronId string) {
	cronKey := fmt.Sprintf("%s:%s", channel, cronId)
	if cronTask, err := p.bot.Memory.Del(p, cronKey); err != nil {
		message := fmt.Sprintf("Invalid cron task '%s'\n%s", cronId, err.Error())
		p.bot.SendMessage(message, channel, p.username, p.icon_url)
	} else {
		if err := p.restartCronTasks(); err != nil {
			message := fmt.Sprintf("Failed to restart cron '%s'\n%s", cronId, err.Error())
			p.bot.SendMessage(message, channel, p.username, p.icon_url)
		} else {
			message := "Deleted cron task.\n"
			message += "```\n"
			message += fmt.Sprintf("%s: %s\n", cronId, cronTask)
			message += "```"
			p.bot.SendMessage(message, channel, p.username, p.icon_url)
		}
	}
}

func (p *Plugin) listCronTasks(channel string) {
	cronList := map[string]string{}

	// get tasks in the specified channel
	if list, err := p.bot.Memory.List(p); err == nil {
		for cronKey, cronTask := range list {
			if strings.HasPrefix(cronKey, channel+":") {
				cronList[cronKey] = cronTask
			}
		}
	}

	if len(cronList) == 0 {
		message := "Could not find cron tasks."
		p.bot.SendMessage(message, channel, p.username, p.icon_url)
	} else {
		message := "```\n"
		for cronKey, cronTask := range cronList {
			cronId := cronKey[len(channel)+1:]
			message += fmt.Sprintf("%s: %s\n", cronId, cronTask)
		}
		message += "```"
		p.bot.SendMessage(message, channel, p.username, p.icon_url)
	}
}

func (p *Plugin) restartCronTasks() error {
	// renew cron client
	if p.cron != nil {
		p.cron.Stop()
	}
	p.cron = cron.New()

	// get all tasks
	re := regexp.MustCompile(`^` + "`" + `\s*([^` + "`" + `]+)\s*` + "`" + `\s+(.+)$`)
	cronList := map[string]string{}
	cronList, _ = p.bot.Memory.List(p)

	// add tasks to the cron client
	for cronKey, cronTask := range cronList {
		channel := cronKey[:strings.Index(cronKey, ":")]

		submatch := re.FindSubmatch([]byte(cronTask))

		p.cron.AddFunc(string(submatch[1]), func() {
			p.bot.SendMessage(string(submatch[2]), channel, p.username, p.icon_url)
		})
	}

	// start scheduler
	p.cron.Start()
	return nil
}
