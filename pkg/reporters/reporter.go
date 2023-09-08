package reporters

import (
	"main/pkg/constants"
	reportPkg "main/pkg/report"
)

type Reporter interface {
	Init()
	Name() constants.ReporterName
	Enabled() bool
	Send(*reportPkg.Report) error
}
