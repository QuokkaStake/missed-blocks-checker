package state

import (
	"main/pkg/events"
	"main/pkg/types"
)

type SnapshotEntry struct {
	OperatorAddress string
	Moniker         string
	Status          int32
	Jailed          bool
	SignatureInfo   types.SignatureInto
}

type Snapshot struct {
	Entries map[string]SnapshotEntry
}

func NewSnapshot(entries map[string]SnapshotEntry) *Snapshot {
	return &Snapshot{Entries: entries}
}

func (snapshot *Snapshot) GetDiff(olderSnapshot *Snapshot) *SnapshotDiff {
	var entries []SnapshotDiffEntry

	for valoper, entry := range snapshot.Entries {
		olderEntry, ok := olderSnapshot.Entries[valoper]
		if !ok {
			continue
		}

		signedBlocksEqual := olderEntry.SignatureInfo.GetNotSigned() != entry.SignatureInfo.GetNotSigned()
		jailedEqual := olderEntry.Jailed == entry.Jailed

		if signedBlocksEqual && jailedEqual {
			entries = append(entries, events.ValidatorGroupChanged{
				Moniker:            entry.Moniker,
				OperatorAddress:    entry.OperatorAddress,
				MissedBlocksBefore: olderEntry.SignatureInfo.GetNotSigned(),
				MissedBlocksAfter:  entry.SignatureInfo.GetNotSigned(),
			})
		}

		if entry.Jailed && !olderEntry.Jailed {
			entries = append(entries, events.ValidatorJailed{
				Moniker:         entry.Moniker,
				OperatorAddress: entry.OperatorAddress,
			})
		}

		if !entry.Jailed && olderEntry.Jailed {
			entries = append(entries, events.ValidatorUnjailed{
				Moniker:         entry.Moniker,
				OperatorAddress: entry.OperatorAddress,
			})
		}
	}

	return &SnapshotDiff{Entries: entries}
}
