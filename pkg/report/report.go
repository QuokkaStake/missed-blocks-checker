package report

type ReportEntry interface {
}

type Report struct {
	Entries []ReportEntry
}

func (d *Report) Empty() bool {
	return len(d.Entries) == 0
}
