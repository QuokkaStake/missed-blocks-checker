package reporters

import (
	"main/pkg/constants"
	statePkg "main/pkg/report"
)

type Reporter interface {
	Init()
	Name() constants.ReporterName
	Enabled() bool
	Send(report *statePkg.Report) error
}
