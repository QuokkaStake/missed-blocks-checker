package types

import (
	"fmt"
	"main/pkg/utils"
)

type Notifier struct {
	OperatorAddress string
	Reporter        string
	Notifier        string
}

func (n Notifier) Equals(another *Notifier) bool {
	return n.OperatorAddress == another.OperatorAddress &&
		n.Reporter == another.Reporter &&
		n.Notifier == another.Notifier
}

type Notifiers []*Notifier

func (n Notifiers) AddNotifier(operatorAddress, reporter, notifier string) (*Notifiers, bool) {
	newNotifier := &Notifier{
		OperatorAddress: operatorAddress,
		Reporter:        reporter,
		Notifier:        notifier,
	}

	fmt.Printf("add notifier %+v\n", n)

	if _, found := utils.Find(n, func(notifier *Notifier) bool {
		return notifier.Equals(newNotifier)
	}); found {
		return &n, false
	}

	newN := append(n, newNotifier)
	return &newN, true
}

func (n Notifiers) GetNotifiers(operatorAddress, reporter string) []string {
	notifiers := utils.Filter(n, func(notifier *Notifier) bool {
		return notifier.OperatorAddress == operatorAddress && notifier.Reporter == reporter
	})

	return utils.Map(notifiers, func(notifier *Notifier) string {
		return notifier.Notifier
	})
}

func (n Notifiers) RemoveNotifier(operatorAddress, reporter, notifier string) (*Notifiers, bool) {
	deletedNotifier := &Notifier{
		OperatorAddress: operatorAddress,
		Reporter:        reporter,
		Notifier:        notifier,
	}

	if _, found := utils.Find(n, func(notifier *Notifier) bool {
		return notifier.Equals(deletedNotifier)
	}); !found {
		return &n, false
	}

	newN := utils.Filter(n, func(notifier *Notifier) bool {
		return !notifier.Equals(deletedNotifier)
	})

	var newNNotifiers Notifiers = newN
	return &newNNotifiers, true
}
