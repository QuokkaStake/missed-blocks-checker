package config

import "fmt"

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
