package main

import (
	"mattermost-bot"
	"mattermost-bot/plugins/batch"
	"mattermost-bot/plugins/cron"
	"mattermost-bot/plugins/echo"
	"mattermost-bot/plugins/help"
	"mattermost-bot/plugins/ping"
)

func main() {
	bot := mmbot.NewBotKit()
	bot.AddPlugin(batch.NewPlugin(bot))
	bot.AddPlugin(cron.NewPlugin(bot))
	bot.AddPlugin(echo.NewPlugin(bot))
	bot.AddPlugin(help.NewPlugin(bot))
	bot.AddPlugin(ping.NewPlugin(bot))
	bot.Run()
}
