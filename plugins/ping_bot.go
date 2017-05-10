package plugins

import (
	"regexp"

	"github.com/mattermost/platform/model"
	"mattermost-bot/botkit"
)

type PingBot struct {
	Bot *botkit.Bot
}

func (p *PingBot) Reply(post *model.Post) error {
	re := regexp.MustCompile(`(?i)(?:^|\W)ping(?:$|\W)`)
	if re.MatchString(post.Message) {
		message := "PONG"
		response := &model.Post{Message: message, ChannelId: post.ChannelId, RootId: post.Id}
		p.Bot.Say(response)
	}

	return nil
}
