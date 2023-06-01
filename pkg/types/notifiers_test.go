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
