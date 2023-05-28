package snapshot

import (
	"main/pkg/config"
	"main/pkg/events"
	"main/pkg/report"
	"main/pkg/types"
)

type Entry struct {
	Validator     *types.Validator
	SignatureInfo types.SignatureInto
}

type Entries map[string]Entry

func (e Entries) ToSlice() []Entry {
	entries := make([]Entry, len(e))

	index := 0
	for _, entry := range e {
		entries[index] = entry
		index++
	}

	return entries
}

type Snapshot struct {
	Entries Entries
}

func (snapshot *Snapshot) GetReport(
	olderSnapshot Snapshot,
	chainConfig *config.ChainConfig,
) (*report.Report, error) {
	var entries []report.Entry

	for valoper, entry := range snapshot.Entries {
		olderEntry, ok := olderSnapshot.Entries[valoper]
		if !ok {
			continue
		}

		missedBlocksBefore := olderEntry.SignatureInfo.GetNotSigned()
		missedBlocksAfter := entry.SignatureInfo.GetNotSigned()

		beforeGroup, err := chainConfig.MissedBlocksGroups.GetGroup(missedBlocksBefore)
		if err != nil {
			return nil, err
		}
		afterGroup, err := chainConfig.MissedBlocksGroups.GetGroup(missedBlocksAfter)
		if err != nil {
			return nil, err
		}

		missedBlocksGroupsEqual := beforeGroup.Start == afterGroup.Start

		if !missedBlocksGroupsEqual && !entry.Validator.Jailed {
			entries = append(entries, events.ValidatorGroupChanged{
				Validator:               entry.Validator,
				MissedBlocksBefore:      missedBlocksBefore,
				MissedBlocksAfter:       missedBlocksAfter,
				MissedBlocksGroupBefore: beforeGroup,
				MissedBlocksGroupAfter:  afterGroup,
			})
		}

		if entry.Validator.SigningInfo != nil &&
			olderEntry.Validator.SigningInfo != nil &&
			entry.Validator.SigningInfo.Tombstoned &&
			!olderEntry.Validator.SigningInfo.Tombstoned {
			entries = append(entries, events.ValidatorTombstoned{
				Validator: entry.Validator,
			})
			continue
		}

		if entry.Validator.Jailed && !olderEntry.Validator.Jailed && olderEntry.Validator.Active() {
			entries = append(entries, events.ValidatorJailed{
				Validator: entry.Validator,
			})
		}

		if !entry.Validator.Jailed && olderEntry.Validator.Jailed {
			entries = append(entries, events.ValidatorUnjailed{
				Validator: entry.Validator,
			})
		}

		if entry.Validator.Active() && !olderEntry.Validator.Active() {
			entries = append(entries, events.ValidatorActive{
				Validator: entry.Validator,
			})
		}

		if !entry.Validator.Active() && olderEntry.Validator.Active() {
			entries = append(entries, events.ValidatorInactive{
				Validator: entry.Validator,
			})
		}
	}

	return &report.Report{Entries: entries}, nil
}

type Info struct {
	Height   int64
	Snapshot Snapshot
}
