package telegram

import (
	"fmt"
	"main/pkg/config"
	"main/pkg/types"
)

type missingValidatorsRender struct {
	Config     *config.ChainConfig
	Validators []missingValidatorsEntry
}

type missingValidatorsEntry struct {
	Validator    *types.Validator
	NotSigned    int64
	Link         types.Link
	BlocksWindow int64
}

func (e missingValidatorsEntry) FormatMissed() string {
	return fmt.Sprintf(
		"%.2f",
		float64(e.NotSigned)/float64(e.BlocksWindow)*100,
	)
}
