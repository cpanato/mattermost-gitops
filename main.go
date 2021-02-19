package main

import (
	"flag"
	"log"
	"os"
	"path"

	"github.com/cpanato/mattermost-gitops/pkg/config"
	"github.com/cpanato/mattermost-gitops/pkg/mattermost"
	"github.com/cpanato/mattermost-gitops/pkg/reconciler"
)

type options struct {
	dryRun                bool
	allowInsecureSHA1     bool
	allowInsecureTLS      bool
	ignoreDefaultChannels bool
	config                string
	authConfig            string
}

func parseOptions() options {
	var o options
	flag.BoolVar(&o.dryRun, "dry-run", true, "does nothing if true (which is the default)")
	flag.BoolVar(&o.ignoreDefaultChannels, "ignore-default-channels", true, "ignore the default channels like Off-Topic and Town Square (default to true)")
	flag.StringVar(&o.config, "config", "", "path to a configuration file, or directory of files")
	flag.StringVar(&o.authConfig, "auth", "auth.json", "path to mattermost auth")
	flag.BoolVar(&o.allowInsecureSHA1, "insecure-sha1-intermediate", false, "allows to use insecure TLS protocols, such as SHA-1")
	flag.BoolVar(&o.allowInsecureTLS, "insecure-tls-version", false, "allows to use TLS versions 1.0 and 1.1")
	flag.Parse()
	return o
}

func main() {
	log.Println("Starting Mattermost GitOps reconciler")
	o := parseOptions()

	c, err := mattermost.LoadConfig(o.authConfig)
	if err != nil {
		log.Fatalf("Failed to load mattermost auth config: %v.\n", err)
	}

	client, serverVersion, err := mattermost.InitClientWithCredentials(c, o.allowInsecureSHA1, o.allowInsecureTLS)
	if err != nil {
		log.Fatalf("Failed to init client: %v\n", err)
	}
	log.Printf("Mattermost Server Version: %s", serverVersion)

	stat, err := os.Stat(o.config)
	if err != nil {
		log.Fatalf("Failed to stat %s: %v\n", o.config, err)
	}
	p := config.NewParser()

	if stat.IsDir() {
		// err = p.ParseDir(o.config)
		// no-op
	} else {
		err = p.ParseFile(o.config, path.Dir(o.config))
	}
	if err != nil {
		log.Fatalf("Failed to load config: %v\n", err)
	}

	strDict := map[string]bool{}
	for _, value := range p.Config.Channels {
		if _, exist := strDict[value.Name]; exist {
			log.Fatalf("Reconciliation failed: Channel name duplicate in config file. Channel name is unique: channel name = %s", value.Name)
		} else {
			strDict[value.Name] = true
		}
	}

	r := reconciler.New(client, p.Config)
	if err := r.Reconcile(o.dryRun, o.ignoreDefaultChannels); err != nil {
		log.Fatalf("Reconciliation failed: %v\n", err)
	}
}
