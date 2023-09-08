package reporters

import (
	"main/pkg/constants"
	reportPkg "main/pkg/types"
)

type Reporter interface {
	Init()
	Name() constants.ReporterName
	Enabled() bool
	Send(*reportPkg.Report) error
}
