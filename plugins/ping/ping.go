package ping

import (
	"regexp"

	"mattermost-bot"
)

type adaptor struct {
	bot *mmbot.BotKit
}

func New(bot *mmbot.BotKit) *mmbot.BotPluginInterface {
	return &adaptor{bot}
}

func (a *adaptor) HandleMessage(command, channel, username string) error {
	re := regexp.MustCompile(`(?i)^ping$`)
	if re.MatchString(command) {
		a.bot.SendMessage("PONG", "", "", "")
	}

	return nil
}

func (a *adaptor) Usage() string {
	return `ping: ping to the bot.`
}
