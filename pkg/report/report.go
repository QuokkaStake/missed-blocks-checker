package report

type Entry interface {
	Type() string
}

type Report struct {
	Entries []Entry
}

func (d *Report) Empty() bool {
	return len(d.Entries) == 0
}
