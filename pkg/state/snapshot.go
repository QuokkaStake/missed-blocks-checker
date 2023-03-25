package state

import (
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

		if olderEntry.SignatureInfo.GetNotSigned() != entry.SignatureInfo.GetNotSigned() {
			entries = append(entries, ValidatorGroupChanged{
				Moniker:            entry.Moniker,
				OperatorAddress:    entry.OperatorAddress,
				MissedBlocksBefore: olderEntry.SignatureInfo.GetNotSigned(),
				MissedBlocksAfter:  entry.SignatureInfo.GetNotSigned(),
			})
		}
	}

	return &SnapshotDiff{Entries: entries}
}
