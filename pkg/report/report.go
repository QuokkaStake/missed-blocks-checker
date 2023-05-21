package report

import "main/pkg/constants"

type Entry interface {
	Type() constants.EventName
}

type Report struct {
	Entries []Entry
}

func (d *Report) Empty() bool {
	return len(d.Entries) == 0
}
