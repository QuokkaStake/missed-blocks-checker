package config

import (
	"errors"
	"fmt"

	"gopkg.in/guregu/null.v4"
)

type ChainPagination struct {
	BlocksSearch   int    `default:"100"  toml:"blocks-search"`
	ValidatorsList uint64 `default:"1000" toml:"validators-list"`
	SigningInfos   uint64 `default:"1000" toml:"signing-infos"`
}
type ChainConfig struct {
	Name                 string          `toml:"name"`
	PrettyName           string          `toml:"pretty-name"`
	RPCEndpoints         []string        `toml:"rpc-endpoints"`
	StoreBlocks          int64           `default:"20000"      toml:"store-blocks"`
	BlocksWindow         int64           `default:"10000"      toml:"blocks-window"`
	MinSignedPerWindow   float64         `default:"0.05"       toml:"min-signed-per-window"`
	QueryEachSigningInfo null.Bool       `default:"false"      toml:"query-each-signing-info"`
	Pagination           ChainPagination `toml:"pagination"`
	Intervals            IntervalsConfig `toml:"intervals"`

	IsConsumer              null.Bool `default:"false"                  toml:"consumer"`
	ProviderRPCEndpoints    []string  `toml:"provider-rpc-endpoints"`
	ConsumerValidatorPrefix string    `toml:"consumer-validator-prefix"`
	ConsumerChainID         string    `toml:"consumer-chain-id"`
	ConsumerSoftOptOut      float64   `default:"0.05"                   toml:"consumer-soft-opt-out"`

	MissedBlocksGroups MissedBlocksGroups `toml:"-"`
	Thresholds         []float64          `default:"[0, 0.5, 1, 5, 10, 25, 50, 75, 90, 100]"                                                    toml:"thresholds"`
	EmojisStart        []string           `default:"[\"🟡\", \"🟡\", \"🟡\", \"🟠\", \"🟠\", \"🟠\", \"🔴\", \"🔴\", \"🔴\"]"                            toml:"emoji-start"`
	EmojisEnd          []string           `default:"[\"🟢\", \"🟡\", \"🟡\", \"🟡\", \"🟡\", \"🟠\", \"🟠\", \"🟠\", \"🟠\"]"                            toml:"emoji-end"`

	ExplorerConfig ExplorerConfig `toml:"explorer"`
	TelegramConfig TelegramConfig `toml:"telegram"`
	DiscordConfig  DiscordConfig  `toml:"discord"`
}

func (c *ChainConfig) GetName() string {
	if c.PrettyName != "" {
		return c.PrettyName
	}

	return c.Name
}

func (c *ChainConfig) GetBlocksSignCount() int64 {
	return int64(float64(c.BlocksWindow) * (1 - c.MinSignedPerWindow))
}

func (c *ChainConfig) GetBlocksMissCount() int64 {
	return int64(float64(c.BlocksWindow) * c.MinSignedPerWindow)
}

func (c *ChainConfig) Validate() error {
	if c.Name == "" {
		return errors.New("chain name is not provided")
	}

	if len(c.RPCEndpoints) == 0 {
		return errors.New("chain has 0 RPC endpoints")
	}

	if len(c.Thresholds) <= 2 {
		return errors.New("not enough thresholds provided")
	}

	if len(c.Thresholds) != len(c.EmojisStart)+1 {
		return fmt.Errorf("got %d start emojis but %d thresholds", len(c.EmojisStart), len(c.Thresholds))
	}

	if len(c.Thresholds) != len(c.EmojisEnd)+1 {
		return fmt.Errorf("got %d end emojis but %d thresholds", len(c.EmojisStart), len(c.Thresholds))
	}

	if c.Thresholds[0] != 0 {
		return fmt.Errorf("first threshold should be 0, but got %.2f", c.Thresholds[0])
	}

	if c.Thresholds[len(c.Thresholds)-1] != 100 {
		return fmt.Errorf("last threshold should be 100, but got %.2f", c.Thresholds[len(c.Thresholds)-1])
	}

	for index, threshold := range c.Thresholds {
		if index == 0 {
			continue
		}

		if threshold <= c.Thresholds[index-1] {
			return fmt.Errorf(
				"threshold at index %d is less than threshold at index %d: %.2f <= %.2f",
				index,
				index-1,
				threshold,
				c.Thresholds[index-1],
			)
		}
	}

	if c.IsConsumer.Bool {
		if len(c.ProviderRPCEndpoints) == 0 {
			return errors.New("chain is a consumer, but has 0 provider RPC endpoints")
		}

		if c.ConsumerChainID == "" {
			return errors.New("chain is a consumer, but consumer chain id is not provided")
		}
	}

	return nil
}

func (c *ChainConfig) RecalculateMissedBlocksGroups() {
	totalRange := float64(c.BlocksWindow) + 1 // from 0 till max blocks allowed, including

	groupsCount := len(c.Thresholds) - 1
	groups := make([]*MissedBlocksGroup, groupsCount)

	for i := 0; i < groupsCount; i++ {
		start := totalRange * c.Thresholds[i] / 100
		end := totalRange*c.Thresholds[i+1]/100 - 1

		groups[i] = &MissedBlocksGroup{
			Start:      int64(start),
			End:        int64(end),
			EmojiStart: c.EmojisStart[i],
			EmojiEnd:   c.EmojisEnd[i],
			DescStart:  fmt.Sprintf("is skipping blocks (> %.1f%%)", c.Thresholds[i]),
			DescEnd:    fmt.Sprintf("is recovering (< %.1f%%)", c.Thresholds[i+1]),
		}
	}

	groups[0].DescEnd = fmt.Sprintf("is recovered (< %.1f%%)", c.Thresholds[1])

	c.MissedBlocksGroups = groups
}
