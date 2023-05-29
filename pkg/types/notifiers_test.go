package types

import (
	"main/pkg/constants"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotifiersEquals(t *testing.T) {
	t.Parallel()

	first := &Notifier{
		OperatorAddress: "address",
		Reporter:        constants.TelegramReporterName,
		Notifier:        "notifier",
	}

	second := &Notifier{
		OperatorAddress: "address",
		Reporter:        constants.TelegramReporterName,
		Notifier:        "notifier",
	}

	assert.True(t, first.Equals(second), "Notifiers should be equal")
}

func TestAddNotifierIfExists(t *testing.T) {
	t.Parallel()

	notifiers := Notifiers{
		&Notifier{
			OperatorAddress: "address",
			Reporter:        constants.TelegramReporterName,
			Notifier:        "notifier",
		},
	}

	newNotifiers, added := notifiers.AddNotifier("address", constants.TelegramReporterName, "notifier")
	assert.False(t, added, "Notifiers should not be added")
	assert.Equal(t, newNotifiers.Length(), 1, "New notifier should not be added!")
}

func TestAddNotifierIfNotExists(t *testing.T) {
	t.Parallel()

	notifiers := Notifiers{
		&Notifier{
			OperatorAddress: "address",
			Reporter:        constants.TelegramReporterName,
			Notifier:        "notifier",
		},
	}

	newNotifiers, added := notifiers.AddNotifier("address", constants.TelegramReporterName, "newnotifier")
	assert.True(t, added, "Notifiers should be added")
	assert.Equal(t, newNotifiers.Length(), 2, "New notifier should be added!")
}

func TestGetNotifiersForReporter(t *testing.T) {
	t.Parallel()

	notifiers := Notifiers{
		&Notifier{
			OperatorAddress: "address",
			Reporter:        constants.TelegramReporterName,
			Notifier:        "notifier1",
		},
		&Notifier{
			OperatorAddress: "address",
			Reporter:        constants.TestReporterName,
			Notifier:        "notifier2",
		},
	}

	reporterNotifiers := notifiers.GetNotifiersForReporter("address", constants.TestReporterName)
	assert.Equal(t, len(reporterNotifiers), 1, "Should have 1 notifier")
	assert.Equal(t, reporterNotifiers[0], "notifier2", "Should have 1 notifier")
}

func TestGetValidatorsForNotifier(t *testing.T) {
	t.Parallel()

	notifiers := Notifiers{
		&Notifier{
			OperatorAddress: "address1",
			Reporter:        constants.TestReporterName,
			Notifier:        "notifier1",
		},
		&Notifier{
			OperatorAddress: "address2",
			Reporter:        constants.TestReporterName,
			Notifier:        "notifier2",
		},
	}

	validatorNotifiers := notifiers.GetValidatorsForNotifier(constants.TestReporterName, "notifier1")
	assert.Len(t, validatorNotifiers, 1, "Should have 1 notifier")
	assert.Equal(t, validatorNotifiers[0], "address1", "Should have 1 notifier")
}

func TestRemoveNotifierIfNotExists(t *testing.T) {
	t.Parallel()

	notifiers := Notifiers{
		&Notifier{
			OperatorAddress: "another_address",
			Reporter:        constants.TelegramReporterName,
			Notifier:        "notifier",
		},
	}

	newNotifiers, removed := notifiers.RemoveNotifier("address", constants.TelegramReporterName, "notifier")
	assert.False(t, removed, "Notifiers should not be removed")
	assert.Equal(t, newNotifiers.Length(), 1, "New notifier should not be removed!")
}

func TestRemoveNotifierIfExists(t *testing.T) {
	t.Parallel()

	notifiers := Notifiers{
		&Notifier{
			OperatorAddress: "address",
			Reporter:        constants.TelegramReporterName,
			Notifier:        "notifier",
		},
	}

	newNotifiers, removed := notifiers.RemoveNotifier("address", constants.TelegramReporterName, "notifier")
	assert.True(t, removed, "Notifiers should be removed")
	assert.Equal(t, newNotifiers.Length(), 0, "New notifier should be removed!")
}
