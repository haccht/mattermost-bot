package echo

import (
	"regexp"

	"mattermost-bot"
)

type Plugin struct {
	bot      *mmbot.BotKit
	username string
	icon_url string
}

func NewPlugin(bot *mmbot.BotKit) *Plugin {
	return &Plugin{bot: bot, username: "Echo"}
}

func (p *Plugin) HandleMessage(text, channel, username string) error {
	re := regexp.MustCompile(`(?i)^echo\s+(.*)$`)
	if re.MatchString(text) {
		bytes := []byte(text)
		group := re.FindSubmatch(bytes)
		p.bot.SendMessage(string(group[1]), channel, p.username, p.icon_url)
	}

	return nil
}

func (p *Plugin) Usage() string {
	return `echo: Echo your message.`
}
