package botkit

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/mattermost/platform/model"
)

type Bot struct {
	client  *model.Client
	adaptor map[string][]BotPlugin
	User    *model.User
	Team    *model.Team
}

type BotPlugin interface {
	Reply(*model.Post) error
}

func NewBot(endpoint, account, password, teamname string) *Bot {
	b := new(Bot)
	b.client = model.NewClient(endpoint)
	b.adaptor = map[string][]BotPlugin{}

	// Ping the server to make sure we can connect.
	if props, err := b.client.GetPing(); err != nil {
		log.Fatalf("There was a problem pinging the Mattermost server.  Are you sure it's running?: %v\n", err.Error())
	} else {
		log.Fatalf("Server detected and is running version %s\n", props["version"])
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

func (b *Bot) Say(post *model.Post) {
	if _, err := b.client.CreatePost(post); err != nil {
		log.Printf("Failed to send a message to the channel: %v\n", err.Error())
	}
}

func (b *Bot) Listen() {
	wsUrl, _ := url.Parse(b.client.Url)
	wsUrl.Scheme = "ws"

	// Lets start listening to some channels via the websocket!
	wsClient, err := model.NewWebSocketClient(wsUrl.String(), b.client.AuthToken)
	if err != nil {
		log.Fatalf("We failed to connect to the websocket: %v\n", err.Error())
	}

	wsClient.Listen()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		for {
			select {
			case resp := <-wsClient.EventChannel:
				b.handleWebsocket(resp)
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

func (b *Bot) Register(plugin BotPlugin, channelName string) {
	if channelResult, err := b.client.GetChannelByName(channelName); err != nil {
		log.Fatalf("Couldn't get channel '%s': %q\n", channelName, err.Error())
	} else {
		channel := channelResult.Data.(*model.Channel)
		plugins := b.adaptor[channel.Id]
		b.adaptor[channel.Id] = append(plugins, plugin)
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

	// Process the post with each plugins
	for channelId, plugins := range b.adaptor {
		if channelId == post.ChannelId {
			for _, plugin := range plugins {
				wg.Add(1)
				go func(post *model.Post, plugin BotPlugin) {
					defer wg.Done()

					err := plugin.Reply(post)
					if err != nil {
						log.Println(err.Error())
					}
				}(post, plugin)
			}
		}
	}

	wg.Wait()
}
