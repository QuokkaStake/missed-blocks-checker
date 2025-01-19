package types

import (
	"main/pkg/constants"
	"time"
)

type HistoricalEvent struct {
	Chain     string
	Type      constants.EventName
	Height    int64
	Validator string
	Event     ReportEvent
	Time      time.Time
}
