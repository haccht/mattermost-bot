package ping

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
	return &Plugin{bot: bot, username: "Ping"}
}

func (p *Plugin) HandleMessage(text, channel, username string) error {
	re := regexp.MustCompile(`(?i)^ping$`)
	if re.MatchString(text) {
		p.bot.SendMessage("PONG", channel, p.username, p.icon_url)
	}

	return nil
}

func (p *Plugin) Usage() string {
	return `ping: See if the bot is alive.`
}
