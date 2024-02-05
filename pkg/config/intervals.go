package config

import "time"

type IntervalsConfig struct {
	Blocks              time.Duration `default:"30"  toml:"blocks"`
	Trim                time.Duration `default:"300" toml:"trim"`
	SlashingParams      time.Duration `default:"300" toml:"slashing-params"`
	SoftOptOutThreshold time.Duration `default:"300" toml:"soft-opt-out-threshold"`
}
