package config

import "time"

type IntervalsConfig struct {
	Blocks         time.Duration `default:"30"  toml:"blocks"`
	LatestBlock    time.Duration `default:"120" toml:"latest_block"`
	Trim           time.Duration `default:"300" toml:"trim"`
	SlashingParams time.Duration `default:"300" toml:"slashing_params"`
}
