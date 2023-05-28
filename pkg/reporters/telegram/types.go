package telegram

import (
	"fmt"
	"main/pkg/config"
	"main/pkg/types"
	"main/pkg/utils"
	"time"
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

type notifierRender struct {
	Config  *config.ChainConfig
	Entries []notifierEntry
}

type notifierEntry struct {
	Link      types.Link
	Notifiers []string
}

type paramsRender struct {
	Config        *config.ChainConfig
	BlockTime     time.Duration
	MaxTimeToJail time.Duration
}

func (r paramsRender) FormatMinSignedPerWindow() string {
	return fmt.Sprintf("%.2f", r.Config.MinSignedPerWindow*100)
}

func (r paramsRender) FormatAvgBlockTime() string {
	return fmt.Sprintf("%.2f", r.BlockTime.Seconds())
}

func (r paramsRender) FormatTimeToJail() string {
	return utils.FormatDuration(r.MaxTimeToJail)
}

func (r paramsRender) FormatGroupPercent(group *config.MissedBlocksGroup) string {
	return fmt.Sprintf(
		"%.2f%% - %.2f%%",
		float64(group.Start)/float64(r.Config.BlocksWindow)*100,
		float64(group.End)/float64(r.Config.BlocksWindow)*100,
	)
}