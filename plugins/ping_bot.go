package plugins

import (
	"encoding/json"
	"regexp"

	"mattermost-bot/botkit"
)

type PingBot struct {
	Bot     *botkit.Bot
	Channel string
}

func (p *PingBot) Reply(post *botkit.Post) error {
	re := regexp.MustCompile(`(?i)(?:^|\W)ping(?:$|\W)`)
	if re.MatchString(post.Message) {
		new_post := map[string]string{
			"username": p.Bot.User.Username,
			"channel":  p.Channel,
			"text":     "PONG",
		}

		b, _ := json.Marshal(new_post)
		p.Bot.PostToWebhook(string(b))

		/*  another way to post message
		new_post := &botkit.Post{Message: response, ChannelId: post.ChannelId, RootId: post.Id}
		p.bot.Post(new_post)
		*/
	}

	return nil
}

func (p *PingBot) ChannelName() string {
	return p.Channel
}
