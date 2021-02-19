package reconciler

import (
	"github.com/cpanato/mattermost-gitops/pkg/config"
	"github.com/mattermost/mattermost-server/v5/model"
)

type channelState struct {
	byName map[string]*config.Channel
}

func (c *channelState) init(m *model.Client4, ignoreDefaultChannels bool) error {
	c.byName = map[string]*config.Channel{}

	var assets []config.Channel
	page := 0
	for {
		channels, resp := m.GetAllChannelsIncludeDeleted(page, 50, "")
		if resp.Error != nil {
			return resp.Error
		}
		if channels.ToJson() == "[]" {
			break
		}

		for _, channel := range *channels {
			if ignoreDefaultChannels {
				if channel.Name == "off-topic" || channel.Name == "town-square" {
					continue
				}
			}

			appendChannel := config.Channel{
				TeamID:      channel.TeamId,
				ChannelID:   channel.Id,
				Private:     !channel.IsOpen(),
				DisplayName: channel.DisplayName,
				Name:        channel.Name,
				Header:      channel.Header,
				Purpose:     channel.Purpose,
			}

			appendChannel.Archive = false
			if channel.DeleteAt != 0 {
				appendChannel.Archive = true
			}

			assets = append(assets, appendChannel)
		}
		page++
	}

	for _, ch := range assets {
		ch2 := ch
		c.byName[ch.Name] = &ch2
	}

	return nil
}

func (c *channelState) update(old string, new config.Channel) error {
	if old != new.Name {
		c.byName[new.Name] = &new
		delete(c.byName, old)
	} else {
		c.byName[old] = &new
	}

	return nil
}
