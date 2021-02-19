package reconciler

import (
	"reflect"
	"testing"

	"github.com/cpanato/mattermost-gitops/pkg/config"
)

func TestReconcileChannels(t *testing.T) {
	tests := []struct {
		name             string
		priorChannels    []config.Channel
		newChannels      []config.Channel
		expectedActions  []Action
		expectedErrCount int
	}{
		{
			name:            "create a new channel",
			priorChannels:   []config.Channel{{Name: "test"}},
			newChannels:     []config.Channel{{Name: "honk"}, {Name: "test"}},
			expectedActions: []Action{createChannelAction{config.Channel{Name: "honk"}}},
		},
		{
			name:          "update a channel",
			priorChannels: []config.Channel{{Name: "honk", DisplayName: "test"}},
			newChannels:   []config.Channel{{Name: "honk", DisplayName: "honk the planet"}},
			expectedActions: []Action{
				updateChannelAction{
					old: config.Channel{
						Name:        "honk",
						DisplayName: "test",
					},
					new: config.Channel{
						Name:        "honk",
						DisplayName: "honk the planet",
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := Reconciler{
				config:   config.Config{Channels: tc.newChannels},
				channels: channelState{byName: map[string]*config.Channel{}},
			}
			for _, c := range tc.priorChannels {
				c2 := c
				r.channels.byName[c.Name] = &c2
			}
			actions, errs := r.reconcileChannels()
			if !reflect.DeepEqual(actions, tc.expectedActions) {
				t.Errorf("Expected actions: %#v\nActual actions: %#v", tc.expectedActions, actions)
			}
			if len(errs) != tc.expectedErrCount {
				t.Errorf("Expected %d errors, but got %d: %v", tc.expectedErrCount, len(errs), errs)
			}
		})
	}
}
