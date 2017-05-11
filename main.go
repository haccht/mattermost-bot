package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"mattermost-bot/botkit"
	"mattermost-bot/plugins"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	account := os.Getenv("MMBOT_ACCOUNT")
	password := os.Getenv("MMBOT_PASSWORD")
	teamname := os.Getenv("MMBOT_TEAMNAME")
	endpoint := os.Getenv("MMBOT_ENDPOINT")

	bot := botkit.NewMMBot(endpoint, account, password, teamname)
	bot.WebhookId = os.Getenv("MMBOT_WEBHOOK")

	if adaptor, _ := plugins.NewPingAdaptor(bot, "Bot_dev"); adaptor != nil {
		bot.Register(adaptor)
	}

	if adaptor, _ := plugins.NewPingAdaptor(bot, "off-topic"); adaptor != nil {
		bot.Register(adaptor)
	}

	if adaptor, _ := plugins.NewRosterAdaptor(bot, "Bot_dev"); adaptor != nil {
		bot.Register(adaptor)
	}

	if adaptor, _ := plugins.NewRosterAdaptor(bot, "magi-TOUBAN"); adaptor != nil {
		bot.Register(adaptor)
	}

	bot.Listen()
}
