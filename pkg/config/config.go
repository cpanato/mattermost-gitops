package config

type Config struct {
	Channels []Channel `json:"channels"`
}

type Channel struct {
	Private     bool   `json:"private"`
	Archive     bool   `json:"archive"`
	TeamID      string `json:"team_id"`
	ChannelID   string `json:"channel_id"`
	DisplayName string `json:"display_name"`
	Name        string `json:"name"`
	Header      string `json:"header"`
	Purpose     string `json:"purpose"`
}
