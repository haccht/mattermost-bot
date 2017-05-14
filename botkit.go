package mmbot

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/mattermost/platform/model"
)

type BotKit struct {
	client  *model.Client
	plugins *PluginManager
	memory  *Memory

	User      *model.User
	Team      *model.Team
	WebhookId string
}

func NewBotKit(endpoint, account, password, teamname string) *BotKit {
	b := new(BotKit)

	b.client = model.NewClient(endpoint)
	b.plugins = NewPluginManager(b)
	b.memory, _ = NewMemory()

	// confirm the mattermost server is alive
	if props, err := b.client.GetPing(); err != nil {
		log.Fatalf("There was a problem pinging the Mattermost server '%s': %v\n", endpoint, err.Error())
	} else {
		log.Printf("Server detected and is running version %s\n", props["version"])
	}

	// login to the mattermost server
	if result, err := b.client.Login(account, password); err != nil {
		log.Fatalf("There was a problem logging into the Mattermost server: %v\n", err.Error())
	} else {
		b.User = result.Data.(*model.User)
	}

	// login to the mattermost team
	if results, err := b.client.GetInitialLoad(); err != nil {
		log.Fatalf("We failed to get the initial load: %v\n", err.Error())
	} else {
		initialLoad := results.Data.(*model.InitialLoad)

		for _, team := range initialLoad.Teams {
			if team.Name == teamname {
				b.Team = team
				break
			}
		}

		if b.Team == nil {
			log.Fatalf("We do not appear to be a member of the team '%s'\n", teamname)
		}

		b.client.SetTeamId(b.Team.Id)
	}

	return b
}

func (b *BotKit) SendMessage(text, channel, username, iconUrl string) error {
	// if the webhook id is not specified, bot will try to send message with api driver
	if b.WebhookId == "" {
		log.Println("Incoming Webhook ID is not set. Try to send message with API driver.")

		var ch *model.Channel
		if result, err := b.client.GetChannelByName(channel); err != nil {
			return fmt.Errorf("Channel '%s' is not found", channel)
		} else {
			ch = result.Data.(*model.Channel)
		}

		post := &model.Post{Message: text, ChannelId: ch.Id}
		err := b.SendMessageWithAPI(post)
		return err
	}

	// default username is the bot user's name logging in.
	if username == "" {
		username = b.User.Username
	}

	// build message
	message := map[string]string{"text": text, "username": username}

	if channel != "" {
		message["channel"] = channel
	}

	if iconUrl != "" {
		message["icon_url"] = iconUrl
	}

	// send message with incoming webhook
	payload, _ := json.Marshal(map[string]map[string]string{"payload": message})
	if _, err := b.client.PostToWebhook(b.WebhookId, string(payload)); err != nil {
		return fmt.Errorf("Failed to send a message: %v", payload, err.Error())
	}

	return nil
}

func (b *BotKit) SendMessageWithAPI(post *model.Post) error {
	// send message with api driver
	if _, err := b.client.CreatePost(post); err != nil {
		return fmt.Errorf("Failed to send a message with api driver: %v", err.Error())
	}

	return nil
}

func (b *BotKit) Run() {
	wsUrl, _ := url.Parse(b.client.Url)
	wsUrl.Scheme = "ws"

	// start listening to websocket
	wsClient, err := model.NewWebSocketClient(wsUrl.String(), b.client.AuthToken)
	if err != nil {
		log.Fatalf("Failed to connect to the websocket '%s': %v\n", wsUrl.String(), err.Error())
	} else {
		log.Printf("Listening to websocket '%s'\n", wsUrl.String())
	}

	wsClient.Listen()
	go func() {
		for {
			select {
			case event := <-wsClient.EventChannel:
				b.handleWebsocketEvent(event)
			}
		}
	}()

	// recieve interruption to stop the bot
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	close := make(chan bool, 1)

	go func() {
		for _ = range sig {
			if wsClient != nil {
				wsClient.Close()
			}

			close <- true
		}
	}()

	<-close
}

func (b *BotKit) AddPlugin(channel string, plugin *BotPlugin) error {
	log.Printf("Add plugin '%v' to channel '%s'", plugin, channel)
	b.plugins.Add(channel, plugin)
	return nil
}

func (b *BotKit) handleWebsocketEvent(event *model.WebSocketEvent) {
	// ignore the event if it is not a posted event
	if event.Event != model.WEBSOCKET_EVENT_POSTED {
		return
	}

	// ignore the post from the bot itself
	post := model.PostFromJson(strings.NewReader(event.Data["post"].(string)))
	if post == nil || post.UserId == b.User.Id {
		return
	}

	var botName, botLinkedName string
	var text, username, channel string

	botName = b.User.Username
	botLinkedName = fmt.Sprintf("<!%s>", b.User.Username)

	// ignore unless the post begins with the bot name or bot linked name
	switch {
	case strings.HasPrefix(post.Message, botName):
		text = post.Message[len(botName):]
	case strings.HasPrefix(post.Message, botLinkedName):
		text = post.Message[len(botLinkedName):]
	default:
		return
	}

	if result, err := b.client.GetChannel(post.ChannelId, ""); err != nil {
		return
	} else {
		ch := result.Data.(*model.Channel)
		channel = ch.Name
	}

	if result, err := b.client.GetUser(post.UserId, ""); err != nil {
		return
	} else {
		user := result.Data.(*model.User)
		username = user.Username
	}

	b.plugins.HandleMessage(text, username, channel)
}
