package telegram

import (
	"fmt"
	"main/pkg/config"
	"main/pkg/types"
	"main/pkg/utils"
	"time"

	tele "gopkg.in/telebot.v3"
)

type Command struct {
	Name    string
	Execute func(c tele.Context) (string, error)
}

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
	Notifiers []*types.Notifier
}

type paramsRender struct {
	Config                   *config.ChainConfig
	BlockTime                time.Duration
	MaxTimeToJail            time.Duration
	ConsumerOptOutValidators int
	ValidatorsCount          int
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

func (r paramsRender) GetConsumerRequiredValidators() int {
	return r.ValidatorsCount - r.ConsumerOptOutValidators
}

func (r paramsRender) FormatSnapshotInterval() string {
	if r.Config.SnapshotsInterval == 1 {
		return "every block"
	}

	return fmt.Sprintf("every %d blocks", r.Config.SnapshotsInterval)
}

func (r paramsRender) FormatSoftOptOut() string {
	return fmt.Sprintf("%.2f", r.Config.ConsumerSoftOptOut*100)
}

type statusEntry struct {
	IsActive    bool
	NeedsToSign bool
	Validator   *types.Validator
	Error       error
	SigningInfo types.SignatureInto
	Link        types.Link
}

type statusRender struct {
	Entries     []statusEntry
	ChainConfig *config.ChainConfig
}

func (s statusRender) FormatNotSignedPercent(entry statusEntry) string {
	return fmt.Sprintf("%.2f", float64(entry.SigningInfo.GetNotSigned())/float64(s.ChainConfig.BlocksWindow)*100)
}

func (s statusRender) FormatVotingPower(entry statusEntry) string {
	text := fmt.Sprintf("%.2f%% VP", entry.Validator.VotingPowerPercent*100)

	if s.ChainConfig.IsConsumer.Bool {
		if entry.NeedsToSign {
			text += ", needs to sign blocks"
		} else {
			text += ", does not need to sign blocks"
		}
	}

	return text
}
