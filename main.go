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

	bot := botkit.NewBot(endpoint, account, password, teamname)

	bot.Register(&plugins.PingBot{bot}, "bot_dev")
	bot.Register(&plugins.PingBot{bot}, "off-topic")
	bot.Listen()
}
