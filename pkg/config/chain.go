package config

import (
	"fmt"
	"gopkg.in/guregu/null.v4"
)

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

func (c *ChainConfig) SetDefaultMissedBlocksGroups() {
	if c.MissedBlocksGroups != nil {
		// GetDefaultLogger().Debug().Msg("MissedBlockGroups is set, not setting the default ones.")
		return
	}

	totalRange := float64(c.BlocksWindow) + 1 // from 0 till max blocks allowed, including

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

	c.MissedBlocksGroups = groups
}
