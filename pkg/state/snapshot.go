package state

import "main/pkg/types"

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
