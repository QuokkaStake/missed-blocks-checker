package config

import (
	"fmt"

	"gopkg.in/guregu/null.v4"
)

type ChainConfig struct {
	Name                 string    `toml:"name"`
	PrettyName           string    `toml:"pretty-name"`
	RPCEndpoints         []string  `toml:"rpc-endpoints"`
	StoreBlocks          int64     `default:"20000"      toml:"store-blocks"`
	BlocksWindow         int64     `default:"10000"      toml:"blocks-window"`
	MinSignedPerWindow   float64   `default:"0.05"       toml:"min-signed-per-window"`
	QueryEachSigningInfo null.Bool `default:"false"      toml:"query-each-signing-info"`
	QuerySlashingParams  null.Bool `default:"true"       toml:"query-slashing-params"`

	IsConsumer              null.Bool `default:"false"                  toml:"consumer"`
	ProviderRPCEndpoints    []string  `toml:"provider-rpc-endpoints"`
	ConsumerValidatorPrefix string    `toml:"consumer-validator-prefix"`
	ConsumerChainID         string    `toml:"consumer-chain-id"`

	MissedBlocksGroups MissedBlocksGroups `toml:"missed-blocks-groups"`
	ExplorerConfig     ExplorerConfig     `toml:"explorer"`
	TelegramConfig     TelegramConfig     `toml:"telegram"`
	DiscordConfig      DiscordConfig      `toml:"discord"`
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

func (c *ChainConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("chain name is not provided")
	}

	if len(c.RPCEndpoints) == 0 {
		return fmt.Errorf("chain has 0 RPC endpoints")
	}

	if c.IsConsumer.Bool {
		if len(c.ProviderRPCEndpoints) == 0 {
			return fmt.Errorf("chain is a consumer, but has 0 provider RPC endpoints")
		}

		if c.ConsumerChainID == "" {
			return fmt.Errorf("chain is a consumer, but consumer chain id is not provided")
		}
	}

	return nil
}

func (c *ChainConfig) RecalculateMissedBlocksGroups() {
	totalRange := float64(c.BlocksWindow) + 1 // from 0 till max blocks allowed, including

	percents := []float64{0, 0.5, 1, 5, 10, 25, 50, 75, 90, 100}
	emojiStart := []string{"🟡", "🟡", "🟡", "🟠", "🟠", "🟠", "🔴", "🔴", "🔴"}
	emojiEnd := []string{"🟢", "🟡", "🟡", "🟡", "🟡", "🟠", "🟠", "🟠", "🟠"}

	groupsCount := len(percents) - 1
	groups := make([]*MissedBlocksGroup, groupsCount)

	for i := 0; i < groupsCount; i++ {
		start := totalRange * percents[i] / 100
		end := totalRange*percents[i+1]/100 - 1

		groups[i] = &MissedBlocksGroup{
			Start:      int64(start),
			End:        int64(end),
			EmojiStart: emojiStart[i],
			EmojiEnd:   emojiEnd[i],
			DescStart:  fmt.Sprintf("is skipping blocks (> %.1f%%)", percents[i]),
			DescEnd:    fmt.Sprintf("is recovering (< %.1f%%)", percents[i+1]),
		}
	}

	groups[0].DescEnd = fmt.Sprintf("is recovered (< %.1f%%)", percents[1])

	c.MissedBlocksGroups = groups
}
