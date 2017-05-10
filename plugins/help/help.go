package help

import (
	"fmt"
	"regexp"

	"mattermost-bot"
)

type Plugin struct {
	bot      *mmbot.BotKit
	username string
	icon_url string
}

func NewPlugin(bot *mmbot.BotKit) *Plugin {
	return &Plugin{bot: bot, username: "Help"}
}

func (p *Plugin) HandleMessage(text, channel, username string) error {
	re := regexp.MustCompile(`(?i)^help$`)
	if re.MatchString(text) {
		message := fmt.Sprintf("What can I do for you?\n```\n%s```", p.bot.Usage())
		p.bot.SendMessage(message, channel, p.username, p.icon_url)
	}

	return nil
}

func (p *Plugin) Usage() string {
	return `help: Display this message.`
}
