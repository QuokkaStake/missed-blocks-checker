package state

type SnapshotDiffEntry interface {
}

type ValidatorGroupChanged struct {
	Moniker            string
	OperatorAddress    string
	MissedBlocksBefore int64
	MissedBlocksAfter  int64
}

type SnapshotDiff struct {
	Entries []SnapshotDiffEntry
}

func (d *SnapshotDiff) Empty() bool {
	return len(d.Entries) == 0
}
