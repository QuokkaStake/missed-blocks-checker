package config

type DiscordConfig struct {
	Guild   string `toml:"guild"`
	Token   string `toml:"token"`
	Channel string `toml:"channel"`
}
