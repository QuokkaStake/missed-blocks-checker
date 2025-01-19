package events

import (
	"main/pkg/constants"
	"main/pkg/types"
)

func MapEventTypesToEvent(eventName constants.EventName) types.ReportEvent {
	eventsMap := map[constants.EventName]types.ReportEvent{
		constants.EventValidatorTombstoned:        &ValidatorTombstoned{},
		constants.EventValidatorJailed:            &ValidatorJailed{},
		constants.EventValidatorUnjailed:          &ValidatorUnjailed{},
		constants.EventValidatorInactive:          &ValidatorInactive{},
		constants.EventValidatorActive:            &ValidatorActive{},
		constants.EventValidatorLeftSignatory:     &ValidatorLeftSignatory{},
		constants.EventValidatorJoinedSignatory:   &ValidatorJoinedSignatory{},
		constants.EventValidatorChangedKey:        &ValidatorChangedKey{},
		constants.EventValidatorChangedMoniker:    &ValidatorChangedMoniker{},
		constants.EventValidatorChangedCommission: &ValidatorChangedCommission{},
		constants.EventValidatorCreated:           &ValidatorCreated{},
		constants.EventValidatorGroupChanged:      &ValidatorGroupChanged{},
	}

	return eventsMap[eventName]
}
