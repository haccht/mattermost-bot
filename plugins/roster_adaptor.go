package plugins

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"mattermost-bot/botkit"
)

const ROSTER_FILE = "roster.txt"

type RosterAdaptor struct {
	bot          *botkit.MMBot
	channelNames []string
	username     string
}

func NewRosterAdaptor(bot *botkit.MMBot, channelNames []string) (botkit.MMBotInterface, error) {
	a := &RosterAdaptor{bot, channelNames, "magi当番"}

	members, _ := a.getMembers()
	if len(members) == 1 {
		members = []string{"kenmotsu", "matuhiro", "shirasu", "hachimura"}
		a.setMembers(members)
	}

	rotate := func() {
		// 2週間おき
		_, weeknum := time.Now().ISOWeek()
		if weeknum%2 != 0 {
			return
		}

		// 当番入替え
		members, _ := a.getMembers()
		members = append(members[1:], members[0])
		a.setMembers(members)

		message := fmt.Sprintf("magi当番交代!\n次の当番は`%s`です", members[0])
		new_post := map[string]string{
			"username": a.username,
			"channel":  a.channelName,
			"text":     message,
		}

		b, _ := json.Marshal(new_post)
		a.bot.PostToWebhook(string(b))
	}

	a.bot.Cron.AddFunc("0 0 9 * * 1", rotate)
	return a, nil
}

func (a *RosterAdaptor) Reply(post *botkit.Post) error {
	re := regexp.MustCompile(`当番は(?:だれ|誰)?(?:\?|？)`)
	if re.MatchString(post.Message) {
		members, err := a.getMembers()
		if err != nil {
			return err
		}

		for _, channelName := range a.channelNames {
			new_post := map[string]string{
				"username": a.username,
				"channel":  channelName,
				"text":     fmt.Sprintf("`@%s`です", members[0]),
			}

			b, _ := json.Marshal(new_post)
			a.bot.PostToWebhook(string(b))
		}
	}

	return nil
}

func (a *RosterAdaptor) ChannelNames() []string {
	return a.channelNames
}

func (a *RosterAdaptor) getMembers() ([]string, error) {
	key := "members"
	val, _ := a.bot.Brain.Get(a, key)

	members := strings.Split(val, ":")
	return members, nil
}

func (a *RosterAdaptor) setMembers(members []string) error {
	key := "members"
	val := strings.Join(members, ":")

	a.bot.Brain.Put(a, key, val)
	return nil
}
