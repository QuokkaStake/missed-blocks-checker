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

type Snapshot struct {
	Entries map[string]Entry
}

func NewSnapshot(entries map[string]Entry) Snapshot {
	return Snapshot{Entries: entries}
}

func (snapshot *Snapshot) GetReport(olderSnapshot Snapshot, chainConfig config.ChainConfig) *report.Report {
	var entries []report.Entry

	for valoper, entry := range snapshot.Entries {
		olderEntry, ok := olderSnapshot.Entries[valoper]
		if !ok {
			continue
		}

		missedBlocksBefore := olderEntry.SignatureInfo.GetNotSigned()
		missedBlocksAfter := entry.SignatureInfo.GetNotSigned()

		beforeGroup, _ := chainConfig.MissedBlocksGroups.GetGroup(missedBlocksBefore)
		afterGroup, _ := chainConfig.MissedBlocksGroups.GetGroup(missedBlocksAfter)

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

		if entry.Validator.Jailed && !olderEntry.Validator.Jailed {
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

	return &report.Report{Entries: entries}
}

type Info struct {
	Height   int64
	Snapshot Snapshot
}
