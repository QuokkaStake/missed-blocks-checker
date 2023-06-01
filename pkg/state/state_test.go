package state

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStateGetAddAndLatestBlock(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.AddBlock(&types.Block{Height: 10})
	assert.Equal(t, state.GetLastBlockHeight(), int64(10), "Height mismatch!")
}

func TestStateSetAndGetValidators(t *testing.T) {
	t.Parallel()

	state := NewState()

	validators := types.ValidatorsMap{
		"address": &types.Validator{OperatorAddress: "address", Moniker: "moniker"},
	}

	state.SetValidators(validators)
	validatorsFromState := state.GetValidators()

	assert.Len(t, validatorsFromState, 1, "Length mismatch!")
	assert.Equal(t, validatorsFromState["address"].Moniker, "moniker", "Validator mismatch!")

	validatorFromState, found := state.GetValidator("address")

	assert.True(t, found, 1, "Validator should be present!")
	assert.Equal(t, validatorFromState.Moniker, "moniker", "Validator mismatch!")
}

func TestAddNotifierIfExists(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.SetNotifiers(&types.Notifiers{
		&types.Notifier{
			OperatorAddress: "address",
			Reporter:        constants.TelegramReporterName,
			Notifier:        "notifier",
		},
	})

	added := state.AddNotifier("address", constants.TelegramReporterName, "notifier")
	assert.False(t, added, "Notifiers should not be added")
	assert.Equal(t, state.notifiers.Length(), 1, "New notifier should not be added!")
}

func TestAddNotifierIfNotExists(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.SetNotifiers(&types.Notifiers{
		&types.Notifier{
			OperatorAddress: "address",
			Reporter:        constants.TelegramReporterName,
			Notifier:        "notifier",
		},
	})

	added := state.AddNotifier("address", constants.TelegramReporterName, "newnotifier")
	assert.True(t, added, "Notifiers should be added")
	assert.Equal(t, state.notifiers.Length(), 2, "New notifier should be added!")
}

func TestGetNotifiersForReporter(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.SetNotifiers(&types.Notifiers{
		&types.Notifier{
			OperatorAddress: "address",
			Reporter:        constants.TelegramReporterName,
			Notifier:        "notifier1",
		},
		&types.Notifier{
			OperatorAddress: "address",
			Reporter:        constants.TestReporterName,
			Notifier:        "notifier2",
		},
	})

	reporterNotifiers := state.GetNotifiersForReporter("address", constants.TestReporterName)
	assert.Equal(t, len(reporterNotifiers), 1, "Should have 1 notifier")
	assert.Equal(t, reporterNotifiers[0], "notifier2", "Should have 1 notifier")
}

func TestGetValidatorsForNotifier(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.SetNotifiers(&types.Notifiers{
		&types.Notifier{
			OperatorAddress: "address1",
			Reporter:        constants.TestReporterName,
			Notifier:        "notifier1",
		},
		&types.Notifier{
			OperatorAddress: "address2",
			Reporter:        constants.TestReporterName,
			Notifier:        "notifier2",
		},
	})

	validatorNotifiers := state.GetValidatorsForNotifier(constants.TestReporterName, "notifier1")
	assert.Len(t, validatorNotifiers, 1, "Should have 1 notifier")
	assert.Equal(t, validatorNotifiers[0], "address1", "Should have 1 notifier")
}

func TestRemoveNotifierIfNotExists(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.SetNotifiers(&types.Notifiers{
		&types.Notifier{
			OperatorAddress: "another_address",
			Reporter:        constants.TelegramReporterName,
			Notifier:        "notifier",
		},
	})

	removed := state.RemoveNotifier("address", constants.TelegramReporterName, "notifier")
	assert.False(t, removed, "Notifiers should not be removed")
	assert.Equal(t, state.notifiers.Length(), 1, "New notifier should not be removed!")
}

func TestRemoveNotifierIfExists(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.SetNotifiers(&types.Notifiers{
		&types.Notifier{
			OperatorAddress: "address",
			Reporter:        constants.TelegramReporterName,
			Notifier:        "notifier",
		},
	})

	removed := state.RemoveNotifier("address", constants.TelegramReporterName, "notifier")
	assert.True(t, removed, "Notifiers should be removed")
	assert.Equal(t, state.notifiers.Length(), 0, "New notifier should be removed!")
}

func TestGetBlockTime(t *testing.T) {
	t.Parallel()

	currentTime := time.Now()

	state := NewState()
	state.AddBlock(&types.Block{
		Height: 10,
		Time:   currentTime.Add(-15 * time.Second),
	})
	state.AddBlock(&types.Block{
		Height: 20,
		Time:   currentTime,
	})

	blockTime := state.GetBlockTime()
	assert.Equal(t, blockTime, 1500*time.Millisecond, "Wrong block time!")
}

func TestGetTimeToJail(t *testing.T) {
	t.Parallel()

	currentTime := time.Now()

	state := NewState()
	state.AddBlock(&types.Block{
		Height: 10,
		Time:   currentTime.Add(-15 * time.Second),
	})
	state.AddBlock(&types.Block{
		Height: 20,
		Time:   currentTime,
	})

	config := &configPkg.ChainConfig{
		BlocksWindow:       100,
		MinSignedPerWindow: 0.1,
	}

	blockTime := state.GetTimeTillJail(config, 50)

	// block_time = 1.5s
	// validator missed 50 blocks and needs to skip 40 more to get jailed
	// 40 * 1.5 = 60s
	assert.Equal(t, blockTime, 60*time.Second, "Wrong jail time!")
}
