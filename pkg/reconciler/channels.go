package reconciler

import (
	"fmt"
	"log"

	"github.com/cpanato/mattermost-gitops/pkg/config"
	"github.com/mattermost/mattermost-server/v5/model"
)

func (r *Reconciler) reconcileChannels() ([]Action, []error) {
	missingChannels := map[string]*config.Channel{}

	var actions []Action
	var errors []error // nolint: prealloc

	for _, c := range r.channels.byName {
		missingChannels[c.Name] = c
	}

	for _, c := range r.config.Channels {
		if o, ok := r.channels.byName[c.Name]; ok {
			if o.DisplayName != c.DisplayName || o.Name != c.Name ||
				o.Purpose != c.Purpose || o.Header != c.Header {
				r.channels.update(o.Name, &c)
				actions = append(actions, &updateChannelAction{old: *o, new: c})
			}

			if c.Private != o.Private {
				actions = append(actions, &updateChannelPrivacyAction{old: *o, new: c})
			}

			if c.Archive != o.Archive {
				if c.Archive {
					actions = append(actions, &archiveChannelAction{channelID: o.ChannelID, update: c})
				} else {
					actions = append(actions, &unarchiveChannelAction{channelID: o.ChannelID, update: c})
				}
			}

			delete(missingChannels, o.Name)
		} else {
			actions = append(actions, &createChannelAction{c})
		}
	}

	for _, o := range missingChannels {
		errors = append(errors, fmt.Errorf("channel %s not referenced in config", o.Name))
	}

	return actions, errors
}

type createChannelAction struct {
	config.Channel
}

func (a *createChannelAction) Describe() string {
	return fmt.Sprintf("Create new channel: %s/%s", a.Name, a.DisplayName)
}

func (a *createChannelAction) Perform(reconciler *Reconciler) error {
	channelType := model.CHANNEL_OPEN
	if a.Private {
		channelType = model.CHANNEL_PRIVATE
	}

	ch := &model.Channel{
		TeamId:      a.TeamID,
		Name:        a.Name,
		DisplayName: a.DisplayName,
		Header:      a.Header,
		Purpose:     a.Purpose,
		Type:        channelType,
		CreatorId:   "",
	}

	channelCreated, resp := reconciler.mattermost.CreateChannel(ch)
	if resp.Error != nil {
		log.Fatalf("Failed to create new channel %s: %v\n", a.Name, resp.Error.Error())
	}

	newChannel := &config.Channel{
		TeamID:      channelCreated.TeamId,
		ChannelID:   channelCreated.Id,
		Private:     channelCreated.IsOpen(),
		DisplayName: channelCreated.DisplayName,
		Name:        channelCreated.Name,
		Header:      channelCreated.Header,
		Purpose:     channelCreated.Purpose,
	}

	reconciler.channels.byName[a.Name] = newChannel
	return nil
}

type unarchiveChannelAction struct {
	channelID string
	update    config.Channel
}

func (a *unarchiveChannelAction) Describe() string {
	return fmt.Sprintf("Unarchive channel: %s", a.update.Name)
}

func (a *unarchiveChannelAction) Perform(reconciler *Reconciler) error {
	_, resp := reconciler.mattermost.RestoreChannel(a.channelID)
	if resp.Error != nil {
		log.Fatalf("Failed to restore channel %s: %v\n", a.update.Name, resp.Error.Error())
	}

	return nil
}

type archiveChannelAction struct {
	channelID string
	update    config.Channel
}

func (a *archiveChannelAction) Describe() string {
	return fmt.Sprintf("Archive channel: %s", a.update.Name)
}

func (a *archiveChannelAction) Perform(reconciler *Reconciler) error {
	_, resp := reconciler.mattermost.DeleteChannel(a.channelID)
	if resp.Error != nil {
		log.Fatalf("Failed to delete channel %s: %v\n", a.update.Name, resp.Error.Error())
	}

	return nil
}

type updateChannelPrivacyAction struct {
	old config.Channel
	new config.Channel
}

func (a *updateChannelPrivacyAction) Describe() string {
	newType := "Public"
	if a.new.Private {
		newType = "Private"
	}

	oldType := "Public"
	if a.old.Private {
		oldType = "Private"
	}

	return fmt.Sprintf("Channel %s privacy mode update from %s to %s", a.new.Name, oldType, newType)
}

func (a *updateChannelPrivacyAction) Perform(reconciler *Reconciler) error {
	channelType := model.CHANNEL_OPEN
	if a.new.Private {
		channelType = model.CHANNEL_PRIVATE
	}
	_, resp := reconciler.mattermost.UpdateChannelPrivacy(a.old.ChannelID, channelType)
	if resp.Error != nil {
		log.Fatalf("Failed to update channel privacy %s: %v\n", a.new.Name, resp.Error.Error())
	}

	return nil
}

type updateChannelAction struct {
	old config.Channel
	new config.Channel
}

func (a *updateChannelAction) Describe() string {
	return fmt.Sprintf("Update channel %s from %+v to %+v", a.old.Name, a.old, a.new)
}

func (a *updateChannelAction) Perform(reconciler *Reconciler) error {
	patch := &model.ChannelPatch{
		DisplayName: &a.new.DisplayName,
		Name:        &a.new.Name,
		Purpose:     &a.new.Purpose,
		Header:      &a.new.Header,
	}
	_, resp := reconciler.mattermost.PatchChannel(a.old.ChannelID, patch)
	if resp.Error != nil {
		log.Fatalf("Failed to update channel %s: %v\n", a.new.Name, resp.Error.Error())
	}

	return nil
}
