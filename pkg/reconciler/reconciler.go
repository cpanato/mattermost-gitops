package reconciler

import (
	"fmt"
	"log"

	"github.com/cpanato/mattermost-gitops/pkg/config"
	"github.com/mattermost/mattermost-server/v5/model"
)

type Reconciler struct {
	mattermost *model.Client4
	config     config.Config
	channels   channelState
}

func New(mattermost *model.Client4, cfg config.Config) *Reconciler {
	return &Reconciler{
		mattermost: mattermost,
		config:     cfg,
		channels:   channelState{},
	}
}

func (r *Reconciler) Reconcile(dryRun, ignoreDefaultChannels bool) error {
	if err := r.channels.init(r.mattermost, ignoreDefaultChannels); err != nil {
		return fmt.Errorf("failed to get initial channel state: %v", err)
	}

	var actions []Action
	var errors []error
	a, e := r.reconcileChannels()
	actions = append(actions, a...)
	errors = append(errors, e...)

	failed := false
	if len(errors) > 0 {
		log.Printf("This configuration cannot be applied against the current reality:")
		failed = true
	}

	for i, e := range errors {
		log.Printf("Error %d: %v.\n", i+1, e)
	}

	if !dryRun && failed {
		dryRun = true
		log.Println("We will not execute anything due to errors, but this what we would've done:")
	} else if dryRun {
		log.Println("In dry run mode so taking no action, but this is what we would've done:")
	}

	if len(actions) > 0 {
		for i, a := range actions {
			log.Printf("Step %d: %s.\n", i+1, a.Describe())
			if !dryRun {
				if err := a.Perform(r); err != nil {
					log.Printf("Failed: %v.\n", err)
				}
			}
		}
	} else {
		log.Println("Nothing to do.")
	}

	if failed {
		return fmt.Errorf("there were configuration errors")
	}

	return nil
}

type Action interface {
	Describe() string
	Perform(reconciler *Reconciler) error
}
