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
)

type Bot struct {
	client   *model.Client
	adaptors map[string][]BotInterface

	User      *model.User
	Team      *model.Team
	WebhookId string
}

type Post model.Post

type BotInterface interface {
	Reply(*Post) error
	ChannelName() string
}

func NewBot(endpoint, account, password, teamname string) *Bot {
	b := new(Bot)
	b.client = model.NewClient(endpoint)
	b.adaptors = map[string][]BotInterface{}

	// Ping the server to make sure we can connect.
	if props, err := b.client.GetPing(); err != nil {
		log.Fatalf("There was a problem pinging the Mattermost server.  Are you sure it's running?: %v\n", err.Error())
	} else {
		log.Printf("Server detected and is running version %s\n", props["version"])
	}

	// lets attempt to login to the Mattermost server as the bot user
	// This will set the token required for all future calls
	// You can get this token with client.AuthToken
	if loginResult, err := b.client.Login(account, password); err != nil {
		log.Fatalf("There was a problem logging into the Mattermost server.  Are you sure ran the setup steps from the README.md?: %v\n", err.Error())
	} else {
		b.User = loginResult.Data.(*model.User)
	}

	// Lets load all the stuff we might need
	if initialLoadResults, err := b.client.GetInitialLoad(); err != nil {
		log.Fatalf("We failed to get the initial load: %v\n", err.Error())
	} else {
		initialLoad := initialLoadResults.Data.(*model.InitialLoad)

		// Lets find our bot team
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
		return fmt.Errorf("Incoming webhook ID has not been set")
	}

	payload := fmt.Sprintf("payload=%s", json)
	if _, err := b.client.PostToWebhook(b.WebhookId, payload); err != nil {
		log.Printf("Failed to send a message to : %s\n%v\n", payload, err.Error())
		return err
	}

	return nil
}

func (b *Bot) Listen() {
	wsUrl, _ := url.Parse(b.client.Url)
	wsUrl.Scheme = "ws"

	// Lets start listening to some channels via the websocket!
	wsClient, err := model.NewWebSocketClient(wsUrl.String(), b.client.AuthToken)
	if err != nil {
		log.Fatalf("Failed to connect to the websocket: %v\n", err.Error())
	} else {
		log.Printf("Listening to websoket: %v", wsUrl.String())
	}

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

	// Graceful stop
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

func (b *Bot) Register(adaptor BotInterface) {
	if channelResult, err := b.client.GetChannelByName(adaptor.ChannelName()); err != nil {
		log.Fatalf("Couldn't get channel '%s': %q\n", adaptor.ChannelName(), err.Error())
	} else {
		channel := channelResult.Data.(*model.Channel)
		adaptors := b.adaptors[channel.Id]
		b.adaptors[channel.Id] = append(adaptors, adaptor)
	}
}

func (b *Bot) handleWebsocket(event *model.WebSocketEvent) {
	// Lets only reponded to message posted events
	if event.Event != model.WEBSOCKET_EVENT_POSTED {
		return
	}

	// Ignore self posted events
	post := model.PostFromJson(strings.NewReader(event.Data["post"].(string)))
	if post == nil || post.UserId == b.User.Id {
		return
	}

	var wg sync.WaitGroup

	// Process the post with each adaptors
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
