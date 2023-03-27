package state

import (
	"main/pkg/events"
	"main/pkg/report"
	"main/pkg/types"
)

type SnapshotEntry struct {
	Validator     *types.Validator
	SignatureInfo types.SignatureInto
}

type Snapshot struct {
	Entries map[string]SnapshotEntry
}

func NewSnapshot(entries map[string]SnapshotEntry) *Snapshot {
	return &Snapshot{Entries: entries}
}

func (snapshot *Snapshot) GetReport(olderSnapshot *Snapshot) *report.Report {
	var entries []report.ReportEntry

	for valoper, entry := range snapshot.Entries {
		olderEntry, ok := olderSnapshot.Entries[valoper]
		if !ok {
			continue
		}

		signedBlocksEqual := olderEntry.SignatureInfo.GetNotSigned() != entry.SignatureInfo.GetNotSigned()
		jailedEqual := olderEntry.Validator.Jailed == entry.Validator.Jailed

		if signedBlocksEqual && jailedEqual {
			entries = append(entries, events.ValidatorGroupChanged{
				Validator:          entry.Validator,
				MissedBlocksBefore: olderEntry.SignatureInfo.GetNotSigned(),
				MissedBlocksAfter:  entry.SignatureInfo.GetNotSigned(),
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
	}

	return &report.Report{Entries: entries}
}
