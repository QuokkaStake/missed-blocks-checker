package reporters

import (
	"main/pkg/constants"
	"main/pkg/types"
)

type Reporter interface {
	Init()
	Start()
	Name() constants.ReporterName
	Enabled() bool
	SerializeEvent(event types.ReportEvent) types.RenderEventItem
	Send(report *types.Report) error
}
