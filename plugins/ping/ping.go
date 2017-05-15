package ping

import (
	"regexp"

	"mattermost-bot"
)

type Plugin struct{}

func (p *Plugin) HandleMessage(bot *mmbot.BotKit, command, channel, username string) error {
	re := regexp.MustCompile(`(?i)^ping$`)
	if re.MatchString(command) {
		bot.SendMessage("PONG", channel, "Ping Response", "")
	}

	return nil
}

func (p *Plugin) Usage() string {
	return `ping: ping to the bot.`
}
