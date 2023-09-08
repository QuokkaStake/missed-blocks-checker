package types

import "time"

type RenderEventItem struct {
	Notifiers     Notifiers
	ValidatorLink Link
	Event         ReportEvent
	TimeToJail    time.Duration
}
