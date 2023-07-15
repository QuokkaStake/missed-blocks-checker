package types

import (
	"main/pkg/constants"
	"main/pkg/utils"
)

type Notifier struct {
	OperatorAddress string
	Reporter        constants.ReporterName
	UserID          string
	UserName        string
}

func (n Notifier) Equals(another *Notifier) bool {
	return n.OperatorAddress == another.OperatorAddress &&
		n.Reporter == another.Reporter &&
		n.UserID == another.UserID
}

type Notifiers []*Notifier

func (n Notifiers) Length() int {
	return len(n)
}

func (n Notifiers) AddNotifier(
	operatorAddress string,
	reporter constants.ReporterName,
	userID string,
	userName string,
) (*Notifiers, bool) {
	newNotifier := &Notifier{
		OperatorAddress: operatorAddress,
		Reporter:        reporter,
		UserID:          userID,
		UserName:        userName,
	}

	if _, found := utils.Find(n, func(notifier *Notifier) bool {
		return notifier.Equals(newNotifier)
	}); found {
		return &n, false
	}

	n = append(n, newNotifier)
	return &n, true
}

func (n Notifiers) GetNotifiersForReporter(
	operatorAddress string,
	reporter constants.ReporterName,
) []*Notifier {
	notifiers := utils.Filter(n, func(notifier *Notifier) bool {
		return notifier.OperatorAddress == operatorAddress && notifier.Reporter == reporter
	})

	return notifiers
}

func (n Notifiers) GetValidatorsForNotifier(
	reporter constants.ReporterName,
	userID string,
) []string {
	notifiers := utils.Filter(n, func(notifierInternal *Notifier) bool {
		return notifierInternal.UserID == userID && notifierInternal.Reporter == reporter
	})

	return utils.Map(notifiers, func(notifier *Notifier) string {
		return notifier.OperatorAddress
	})
}

func (n Notifiers) RemoveNotifier(
	operatorAddress string,
	reporter constants.ReporterName,
	userID string,
) (*Notifiers, bool) {
	deletedNotifier := &Notifier{
		OperatorAddress: operatorAddress,
		Reporter:        reporter,
		UserID:          userID,
	}

	if _, found := utils.Find(n, func(notifier *Notifier) bool {
		return notifier.Equals(deletedNotifier)
	}); !found {
		return &n, false
	}

	newN := utils.Filter(n, func(notifier *Notifier) bool {
		return !notifier.Equals(deletedNotifier)
	})

	var newNotifiers Notifiers = newN
	return &newNotifiers, true
}
