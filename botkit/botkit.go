package botkit

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/mattermost/platform/model"
	"github.com/robfig/cron"
)

type Bot struct {
	client    *model.Client
	adaptors  map[string][]BotInterface
	Cron      *cron.Cron
	User      *model.User
	Team      *model.Team
	WebhookId string
}

type BotInterface interface {
	Reply(*Post) error
	ChannelName() string
}

type Post model.Post

func NewBot(endpoint, account, password, teamname string) *Bot {
	b := new(Bot)
	b.client = model.NewClient(endpoint)
	b.adaptors = map[string][]BotInterface{}
	b.Cron = cron.New()

	if props, err := b.client.GetPing(); err != nil {
		log.Fatalf("There was a problem pinging the Mattermost server '%s': %v\n", endpoint, err.Error())
	} else {
		log.Printf("Server detected and is running version %s\n", props["version"])
	}

	if loginResult, err := b.client.Login(account, password); err != nil {
		log.Fatalf("There was a problem logging into the Mattermost server: %v\n", err.Error())
	} else {
		b.User = loginResult.Data.(*model.User)
	}

	if initialLoadResults, err := b.client.GetInitialLoad(); err != nil {
		log.Fatalf("We failed to get the initial load: %v\n", err.Error())
	} else {
		initialLoad := initialLoadResults.Data.(*model.InitialLoad)

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

func (b *Bot) Post(post *Post) error {
	new_post := (*model.Post)(post)
	if _, err := b.client.CreatePost(new_post); err != nil {
		log.Printf("Failed to send a message: %v\n", err.Error())
		return err
	}

	return nil
}

func (b *Bot) PostToWebhook(json string) error {
	if b.WebhookId == "" {
		return fmt.Errorf("Incoming webhook ID has not been set.")
	}

	payload := fmt.Sprintf("payload=%s", json)
	if _, err := b.client.PostToWebhook(b.WebhookId, payload); err != nil {
		log.Printf("Failed to send a message: %v\n", payload, err.Error())
		return err
	}

	return nil
}

func (b *Bot) Listen() {
	wsUrl, _ := url.Parse(b.client.Url)
	wsUrl.Scheme = "ws"

	wsClient, err := model.NewWebSocketClient(wsUrl.String(), b.client.AuthToken)
	if err != nil {
		log.Fatalf("Failed to connect to the websocket '%s': %v\n", wsUrl.String(), err.Error())
	} else {
		log.Printf("Listening to websocket '%s'\n", wsUrl.String())
	}

	b.Cron.Start()
	wsClient.Listen()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		for {
			select {
			case event := <-wsClient.EventChannel:
				b.handleWebsocket(event)
			}
		}
	}()

	close := make(chan bool, 1)

	go func() {
		for _ = range sig {
			b.Cron.Stop()
			if wsClient != nil {
				wsClient.Close()
			}

			close <- true
		}
	}()

	<-close
}

func (b *Bot) Register(adaptor BotInterface) {
	channelName := strings.ToLower(adaptor.ChannelName())
	if channelResult, err := b.client.GetChannelByName(channelName); err != nil {
		log.Fatalf("Couldn't get channel '%s': %v\n", channelName, err.Error())
	} else {
		channel := channelResult.Data.(*model.Channel)
		adaptors := b.adaptors[channel.Id]
		b.adaptors[channel.Id] = append(adaptors, adaptor)
	}
}

func (b *Bot) handleWebsocket(event *model.WebSocketEvent) {
	if event.Event != model.WEBSOCKET_EVENT_POSTED {
		return
	}

	post := model.PostFromJson(strings.NewReader(event.Data["post"].(string)))
	if post == nil || post.UserId == b.User.Id {
		return
	}

	var wg sync.WaitGroup
	for channelId, adaptors := range b.adaptors {
		if channelId == post.ChannelId {
			for _, adaptor := range adaptors {
				wg.Add(1)
				go func(post *model.Post, adaptor BotInterface) {
					defer wg.Done()

					err := adaptor.Reply((*Post)(post))
					if err != nil {
						log.Println(err.Error())
					}
				}(post, adaptor)
			}
		}
	}

	wg.Wait()
}
