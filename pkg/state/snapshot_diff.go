package state

type SnapshotDiffEntry interface {
}

type SnapshotDiff struct {
	Entries []SnapshotDiffEntry
}

func (d *SnapshotDiff) Empty() bool {
	return len(d.Entries) == 0
}
