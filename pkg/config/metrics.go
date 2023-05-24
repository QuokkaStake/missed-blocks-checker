package config

import "gopkg.in/guregu/null.v4"

type MetricsConfig struct {
	Enabled    null.Bool `toml:"enabled" default:"true"`
	ListenAddr string    `toml:"listen-addr" default:":9570"`
}
