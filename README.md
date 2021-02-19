# Mattermost-GitOps

This tool is heavily inspired on [tempelis](https://github.com/kubernetes-sigs/slack-infra/tree/master/tempelis)

It syncronizes the configuration described in a YAML file against your Mattermost installation.
Combined with a CI system, it can be used to implement GitOps for Mattermost.

At this stage, it can:

### Channels
  - Create new channels
  - Update a Channel (Header, Purpose, Channel Display name and Channel name)
  - Update Channel Privacy (Open Channel / Private Channel)
  - Archive/UnArchive Channel


## Config

### Authentication

It expects a config file in the location given by `--auth` that looks like this:

```
{
    "authToken": "ic3hu6ydebbsib1yd7x5wn1nro",
    "instanceUrl": "http://localhost:8065/"
}
```

`authToken` is a value provided by your Mattermost installation, see how to
create a [Personal Token](https://docs.mattermost.com/developer/personal-access-tokens.html?#personal-access-tokens)

`instanceUrl` is your Mattermost URL


#### Channels

It expects a complete list of public channels to be provided. If a public channel exists on
Mattermost that is not in the yaml channel list, it will error out.

A channel list with a single fully-specified channel looks like this:


```yaml
channels:
- team_id: utaq935c5j8z5x3gwske8bep7c # The team ID where you want to create the channel, a Mattermost installation can have multiple teams
  private: false # If a channel is public or private
  display_name: My honk channel # Diplay name in the UI
  name: honk # Channel name
  header: honk the planet # The header for the channel. Optional
  purpose: just to honk # Purpose of the channel. Optional
```


## Future Work

Add support:

- Configuration
- Users
- Webhooks?
