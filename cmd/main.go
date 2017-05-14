package main

import (
	"log"
	"os"

	"mattermost-bot"
	"mattermost-bot/plugins/ping"

	"github.com/joho/godotenv"
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

	bot := mmbot.NewBotKit(endpoint, account, password, teamname)
	bot.WebhookId = os.Getenv("MMBOT_WEBHOOK")

	bot.AddPlugin("Bot_dev", &ping.PingPlugin{})
	bot.Run()
}
