package config

import "time"

type IntervalsConfig struct {
	Blocks         time.Duration `toml:"blocks" default:"30"`
	LatestBlock    time.Duration `toml:"latest_block" default:"120"`
	Trim           time.Duration `toml:"trim" default:"300"`
	SlashingParams time.Duration `toml:"slashing_params" default:"300"`
}
