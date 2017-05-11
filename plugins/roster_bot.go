package plugins

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"

	"mattermost-bot/botkit"
)

type RosterAdaptor struct {
	bot         *botkit.Bot
	channelName string
}

func NewRosterAdaptor(bot *botkit.Bot, channelName string) (botkit.BotInterface, error) {
	return &RosterAdaptor{bot, channelName}, nil
}

func (a *RosterAdaptor) Reply(post *botkit.Post) error {
	re := regexp.MustCompile(`当番は(?:だれ|誰)?(?:\?|？)`)
	if re.MatchString(post.Message) {
		var roster []string

		if fp, err := os.Open("roster.txt"); err != nil {
			log.Println("There exist no 'roster.txt' in the current directory")
			return err
		} else {
			scanner := bufio.NewScanner(fp)
			for scanner.Scan() {
				roster = append(roster, scanner.Text())
			}
			fp.Close()
		}

		new_post := map[string]string{
			"username": "magi当番",
			"channel":  a.channelName,
			"text":     fmt.Sprintf("`@%s` デス", roster[0]),
		}

		b, _ := json.Marshal(new_post)
		a.bot.PostToWebhook(string(b))
	}

	return nil
}

func (a *RosterAdaptor) ChannelName() string {
	return a.channelName
}
