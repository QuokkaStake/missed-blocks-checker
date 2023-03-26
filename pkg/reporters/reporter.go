package reporters

import (
	statePkg "main/pkg/report"
)

type Reporter interface {
	Init()
	Name() string
	Enabled() bool
	Send(report *statePkg.Report) error
}
