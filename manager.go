package mmbot

import (
	"sync"
)

type PluginManager struct {
	bot     *BotKit
	plugins []*BotPluginInterface
}

type BotPluginInterface interface {
	HandleMessage(text, username, channel string) error
	Usage()
}

func NewPluginManager(b *BotKit) *PluginManager {
	m := &PluginManager{}
	m.bot = b
	m.plugins = []*BotPluginInterface{}

	return m
}

func (m *PluginManager) Add(channel string, adaptor BotPluginInterface) {
	m.plugins = append(m.plugins, adaptor)
}

func (m *PluginManager) HandleMessage(text, username, channel string) {
	wg := &sync.WaitGroup{}
	for _, adaptor := range m.plugins {
		if adaptor.Channel == channel {
			wg.Add(1)
			go func(a *BotPluginInterface) {
				defer wg.Done()
				a.HandleMessage(text, username, channel)
			}(adaptor)
		}
	}
	wg.Wait()
}
