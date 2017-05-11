package plugins

import (
	"encoding/json"
	"regexp"

	"mattermost-bot/botkit"
)

type PingAdaptor struct {
	bot         *botkit.Bot
	channelName string
}

func NewPingAdaptor(bot *botkit.Bot, channelName string) (botkit.BotInterface, error) {
	a := &PingAdaptor{bot, channelName}

	// you may schedule a job to execute
	/*
		new_post := map[string]string{
			"username": a.bot.User.Username,
			"channel":  a.channelName,
			"text":     "Hello, are you stil alive?",
		}

		b, _ := json.Marshal(new_post)
		a.bot.Cron.AddFunc("0 * * * * *", func() { a.bot.PostToWebhook(string(b)) })
	*/

	return a, nil
}

func (a *PingAdaptor) Reply(post *botkit.Post) error {
	re := regexp.MustCompile(`(?i)(?:^|\W)ping(?:$|\W)`)
	if re.MatchString(post.Message) {
		new_post := map[string]string{
			"username": a.bot.User.Username,
			"channel":  a.channelName,
			"text":     "PONG",
		}

		b, _ := json.Marshal(new_post)
		a.bot.PostToWebhook(string(b))

		//  another way to post message
		/*
			new_post := &botkit.Post{Message: response, ChannelId: post.ChannelId, RootId: post.Id}
			a.bot.Post(new_post)
		*/
	}

	return nil
}

func (a *PingAdaptor) ChannelName() string {
	return a.channelName
}
