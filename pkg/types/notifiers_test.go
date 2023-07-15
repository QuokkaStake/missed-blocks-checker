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
		UserName:        "notifier",
		UserID:          "id",
	}

	second := &Notifier{
		OperatorAddress: "address",
		Reporter:        constants.TelegramReporterName,
		UserName:        "notifier",
		UserID:          "id",
	}

	assert.True(t, first.Equals(second), "Notifiers should be equal")
}
