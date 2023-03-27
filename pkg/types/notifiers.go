package types

import "main/pkg/utils"

// Notifiers has the following struct:
// notifiers[valoper][reporter] == [reporter1, reporter2]
type Notifiers map[string]map[string][]string

func (n Notifiers) AddNotifier(operatorAddress, reporter, notifier string) bool {
	if _, ok := n[operatorAddress]; !ok {
		n[operatorAddress] = make(map[string][]string)
	}

	if _, ok := n[operatorAddress][reporter]; !ok {
		n[operatorAddress][reporter] = make([]string, 1)
	}

	if utils.Contains(n[operatorAddress][reporter], notifier) {
		return false
	}

	n[operatorAddress][reporter] = append(n[operatorAddress][reporter], notifier)
	return true
}

func (n Notifiers) GetNotifiers(operatorAddress, reporter string) []string {
	if _, ok := n[operatorAddress]; !ok {
		return []string{}
	}

	if _, ok := n[operatorAddress][reporter]; !ok {
		return []string{}
	}

	return n[operatorAddress][reporter]
}
