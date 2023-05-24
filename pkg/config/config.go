package config

import (
	"fmt"
	"main/pkg/types"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/creasty/defaults"
	"gopkg.in/guregu/null.v4"
)

type Config struct {
	LogConfig      LogConfig      `toml:"log"`
	ChainConfigs   []*ChainConfig `toml:"chains"`
	DatabaseConfig DatabaseConfig `toml:"database"`
	MetricsConfig  MetricsConfig  `toml:"metrics"`
}

type MissedBlocksGroup struct {
	Start      int64  `toml:"start"`
	End        int64  `toml:"end"`
	EmojiStart string `toml:"emoji-start"`
	EmojiEnd   string `toml:"emoji-end"`
	DescStart  string `toml:"desc-start"`
	DescEnd    string `toml:"desc-end"`
}

type MissedBlocksGroups []MissedBlocksGroup

// Validate checks that MissedBlocksGroup is an array of sorted MissedBlocksGroup
// covering each interval.
// Example (start - end), given that window = 300:
// 0 - 99, 100 - 199, 200 - 300 - valid
// 0 - 50 - not valid.
func (g MissedBlocksGroups) Validate(window int64) error {
	if len(g) == 0 {
		return fmt.Errorf("MissedBlocksGroups is empty")
	}

	if g[0].Start != 0 {
		return fmt.Errorf("first MissedBlocksGroup's start should be 0, got %d", g[0].Start)
	}

	if g[len(g)-1].End < window {
		return fmt.Errorf("last MissedBlocksGroup's end should be >= %d, got %d", window, g[len(g)-1].End)
	}

	for i := 0; i < len(g)-1; i++ {
		if g[i+1].Start-g[i].End != 1 {
			return fmt.Errorf(
				"MissedBlocksGroup at index %d ends at %d, and the next one starts with %d",
				i,
				g[i].End,
				g[i+1].Start,
			)
		}
	}

	return nil
}

func (g MissedBlocksGroups) GetGroup(missed int64) (*MissedBlocksGroup, error) {
	for _, group := range g {
		if missed >= group.Start && missed <= group.End {
			return &group, nil
		}
	}

	return nil, fmt.Errorf("could not find a group for missed blocks counter = %d", missed)
}

type DatabaseConfig struct {
	Path string `toml:"path"`
}

func (c *DatabaseConfig) Validate() error {
	if c.Path == "" {
		return fmt.Errorf("database path not specified")
	}

	return nil
}

type ExplorerConfig struct {
	ValidatorLinkPattern string `toml:"validator-link-pattern"`
	MintscanPrefix       string `toml:"mintscan-prefix"`
}

func (c *ExplorerConfig) GetValidatorLink(validator *types.Validator) types.Link {
	if c.MintscanPrefix != "" {
		return types.Link{
			Href: fmt.Sprintf(
				"https://mintscan.io/%s/validators/%s",
				c.MintscanPrefix,
				validator.OperatorAddress,
			),
			Text: validator.Moniker,
		}
	}

	if c.ValidatorLinkPattern != "" {
		return types.Link{
			Href: fmt.Sprintf(c.ValidatorLinkPattern, validator.OperatorAddress),
			Text: validator.Moniker,
		}
	}

	return types.Link{Text: validator.Moniker}
}

type ChainConfig struct {
	Name                 string    `toml:"name"`
	RPCEndpoints         []string  `toml:"rpc-endpoints"`
	StoreBlocks          int64     `toml:"store-blocks" default:"20000"`
	BlocksWindow         int64     `toml:"blocks-window" default:"10000"`
	MinSignedPerWindow   float64   `toml:"min-signed-per-window" default:"0.05"`
	QueryEachSigningInfo null.Bool `toml:"query-each-signing-info" default:"false"`

	MissedBlocksGroups MissedBlocksGroups `toml:"missed-blocks-groups"`
	ExplorerConfig     ExplorerConfig     `toml:"explorer"`
	TelegramConfig     TelegramConfig     `toml:"telegram"`
}

func (c *ChainConfig) GetBlocksSignCount() int64 {
	return int64(float64(c.BlocksWindow) * (1 - c.MinSignedPerWindow))
}

func (c *ChainConfig) Validate() error {
	if len(c.RPCEndpoints) == 0 {
		return fmt.Errorf("chain has 0 RPC endpoints")
	}

	return nil
}

type TelegramConfig struct {
	Chat   int64   `toml:"chat"`
	Token  string  `toml:"token"`
	Admins []int64 `toml:"admins"`
}

type LogConfig struct {
	LogLevel   string `toml:"level" default:"info"`
	JSONOutput bool   `toml:"json" default:"false"`
}

func (config *Config) Validate() error {
	if len(config.ChainConfigs) == 0 {
		return fmt.Errorf("no chains specified")
	}

	for index, chainConfig := range config.ChainConfigs {
		if err := chainConfig.Validate(); err != nil {
			return fmt.Errorf("error in chain config #%d: %s", index, err)
		}
	}

	if err := config.DatabaseConfig.Validate(); err != nil {
		return fmt.Errorf("error in database config: %s", err)
	}

	return nil
}

func GetConfig(path string) (*Config, error) {
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	configString := string(configBytes)

	configStruct := &Config{}
	if _, err = toml.Decode(configString, configStruct); err != nil {
		return nil, err
	}
	if err := defaults.Set(configStruct); err != nil {
		return nil, err
	}

	return configStruct, nil
}

func (config *ChainConfig) SetDefaultMissedBlocksGroups() {
	if config.MissedBlocksGroups != nil {
		// GetDefaultLogger().Debug().Msg("MissedBlockGroups is set, not setting the default ones.")
		return
	}

	totalRange := float64(config.BlocksWindow) + 1 // from 0 till max blocks allowed, including

	percents := []float64{0, 0.5, 1, 5, 10, 25, 50, 75, 90, 100}
	emojiStart := []string{"🟡", "🟡", "🟡", "🟠", "🟠", "🟠", "🔴", "🔴", "🔴"}
	emojiEnd := []string{"🟢", "🟡", "🟡", "🟡", "🟡", "🟠", "🟠", "🟠", "🟠"}

	groupsCount := len(percents) - 1
	groups := make([]MissedBlocksGroup, groupsCount)

	for i := 0; i < groupsCount; i++ {
		start := totalRange * percents[i] / 100
		end := totalRange*percents[i+1]/100 - 1

		groups[i] = MissedBlocksGroup{
			Start:      int64(start),
			End:        int64(end),
			EmojiStart: emojiStart[i],
			EmojiEnd:   emojiEnd[i],
			DescStart:  fmt.Sprintf("is skipping blocks (> %.1f%%)", percents[i]),
			DescEnd:    fmt.Sprintf("is recovering (< %.1f%%)", percents[i+1]),
		}
	}

	groups[0].DescEnd = fmt.Sprintf("is recovered (< %.1f%%)", percents[1])

	config.MissedBlocksGroups = groups
}

type MetricsConfig struct {
	Enabled    null.Bool `toml:"enabled" default:"true"`
	ListenAddr string    `toml:"listen-addr" default:":9570"`
}
