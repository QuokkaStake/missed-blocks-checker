package state

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/types"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestStateGetAddAndLatestBlock(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.AddBlock(&types.Block{Height: 10})
	assert.Equal(t, int64(10), state.GetLastBlockHeight(), "Height mismatch!")
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
	assert.Equal(t, "moniker", validatorsFromState["address"].Moniker, "Validator mismatch!")

	validatorFromState, found := state.GetValidator("address")

	assert.True(t, found, "Validator should be present!")
	assert.Equal(t, "moniker", validatorFromState.Moniker, "Validator mismatch!")
}

func TestAddNotifierIfExists(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.SetNotifiers(&types.Notifiers{
		&types.Notifier{
			OperatorAddress: "address",
			Reporter:        constants.TelegramReporterName,
			UserName:        "notifier",
			UserID:          "id",
		},
	})

	added := state.AddNotifier("address", constants.TelegramReporterName, "id", "notifier")

	assert.False(t, added, "Notifiers should not be added")
	assert.Equal(t, 1, state.notifiers.Length(), "New notifier should not be added!")
}

func TestAddNotifierIfNotExists(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.SetNotifiers(&types.Notifiers{
		&types.Notifier{
			OperatorAddress: "address",
			Reporter:        constants.TelegramReporterName,
			UserName:        "notifier",
			UserID:          "id",
		},
	})

	added := state.AddNotifier("address", constants.TelegramReporterName, "id2", "newnotifier")
	assert.True(t, added, "Notifiers should be added")
	assert.Equal(t, 2, state.notifiers.Length(), "New notifier should be added!")
}

func TestGetNotifiersForReporter(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.SetNotifiers(&types.Notifiers{
		&types.Notifier{
			OperatorAddress: "address",
			Reporter:        constants.TelegramReporterName,
			UserName:        "notifier1",
			UserID:          "id1",
		},
		&types.Notifier{
			OperatorAddress: "address",
			Reporter:        constants.TestReporterName,
			UserName:        "notifier2",
			UserID:          "id2",
		},
	})

	reporterNotifiers := state.GetNotifiersForReporter("address", constants.TestReporterName)
	assert.Len(t, reporterNotifiers, 1, "Should have 1 notifier")
	assert.Equal(t, "notifier2", reporterNotifiers[0].UserName, "Should have 1 notifier")
}

func TestGetValidatorsForNotifier(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.SetNotifiers(&types.Notifiers{
		&types.Notifier{
			OperatorAddress: "address1",
			Reporter:        constants.TestReporterName,
			UserName:        "notifier1",
			UserID:          "id1",
		},
		&types.Notifier{
			OperatorAddress: "address2",
			Reporter:        constants.TestReporterName,
			UserName:        "notifier2",
			UserID:          "id2",
		},
	})

	validatorNotifiers := state.GetValidatorsForNotifier(constants.TestReporterName, "id1")
	assert.Len(t, validatorNotifiers, 1, "Should have 1 notifier")
	assert.Equal(t, "address1", validatorNotifiers[0], "Should have 1 notifier")
}

func TestRemoveNotifierIfNotExists(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.SetNotifiers(&types.Notifiers{
		&types.Notifier{
			OperatorAddress: "another_address",
			Reporter:        constants.TelegramReporterName,
			UserName:        "notifier",
			UserID:          "id",
		},
	})

	removed := state.RemoveNotifier("address", constants.TelegramReporterName, "id")
	assert.False(t, removed, "Notifiers should not be removed")
	assert.Equal(t, 1, state.notifiers.Length(), "New notifier should not be removed!")
}

func TestRemoveNotifierIfExists(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.SetNotifiers(&types.Notifiers{
		&types.Notifier{
			OperatorAddress: "address",
			Reporter:        constants.TelegramReporterName,
			UserName:        "notifier",
			UserID:          "id",
		},
	})

	removed := state.RemoveNotifier("address", constants.TelegramReporterName, "id")
	assert.True(t, removed, "Notifiers should be removed")
	assert.Equal(t, 0, state.notifiers.Length(), "New notifier should be removed!")
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
	assert.Equal(t, 1500*time.Millisecond, blockTime, "Wrong block time!")
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
	assert.Equal(t, 60*time.Second, blockTime, "Wrong jail time!")
}

func TestValidatorsMissedBlocksNoBlocks(t *testing.T) {
	t.Parallel()
	state := NewState()
	validator := &types.Validator{}
	signature, err := state.GetValidatorMissedBlocks(validator, 5)

	require.Error(t, err, "Error should be present!")
	assert.Equal(t, int64(0), signature.Signed, "Argument mismatch!")
	assert.Equal(t, int64(0), signature.NoSignature, "Argument mismatch!")
	assert.Equal(t, int64(0), signature.NotSigned, "Argument mismatch!")
	assert.Equal(t, int64(0), signature.NotActive, "Argument mismatch!")
}

func TestValidatorsMissedBlocksAllSigned(t *testing.T) {
	t.Parallel()

	validator := &types.Validator{ConsensusAddressHex: "address"}
	state := NewState()

	state.AddBlock(&types.Block{Height: 1, Signatures: map[string]int32{"address": 3}, Proposer: "address", Validators: map[string]bool{"address": true}})
	state.AddBlock(&types.Block{Height: 2, Signatures: map[string]int32{"address": 3}, Validators: map[string]bool{"address": true}})
	state.AddBlock(&types.Block{Height: 3, Signatures: map[string]int32{"address": 3}, Validators: map[string]bool{"address": true}})
	state.AddBlock(&types.Block{Height: 4, Signatures: map[string]int32{"address": 3}, Validators: map[string]bool{"address": true}})
	state.AddBlock(&types.Block{Height: 5, Signatures: map[string]int32{"address": 3}, Validators: map[string]bool{"address": true}})

	signature, err := state.GetValidatorMissedBlocks(validator, 5)

	require.NoError(t, err, "Error should not be present!")
	assert.Equal(t, int64(5), signature.Signed, "Argument mismatch!")
	assert.Equal(t, int64(0), signature.NoSignature, "Argument mismatch!")
	assert.Equal(t, int64(0), signature.NotSigned, "Argument mismatch!")
	assert.Equal(t, int64(0), signature.NotActive, "Argument mismatch!")
	assert.Equal(t, int64(1), signature.Proposed, "Argument mismatch!")
}

func TestValidatorsMissedBlocksAllMissed(t *testing.T) {
	t.Parallel()

	validator := &types.Validator{ConsensusAddressHex: "address"}
	state := NewState()

	state.AddBlock(&types.Block{Height: 1, Signatures: map[string]int32{}, Validators: map[string]bool{"address": true}})
	state.AddBlock(&types.Block{Height: 2, Signatures: map[string]int32{}, Validators: map[string]bool{"address": true}})
	state.AddBlock(&types.Block{Height: 3, Signatures: map[string]int32{}, Validators: map[string]bool{"address": true}})
	state.AddBlock(&types.Block{Height: 4, Signatures: map[string]int32{"address": 1}, Validators: map[string]bool{"address": true}})
	state.AddBlock(&types.Block{Height: 5, Signatures: map[string]int32{"address": 1}, Validators: map[string]bool{"address": true}})

	signature, err := state.GetValidatorMissedBlocks(validator, 5)

	require.NoError(t, err, "Error should not be present!")
	assert.Equal(t, int64(0), signature.Signed, "Argument mismatch!")
	assert.Equal(t, int64(3), signature.NoSignature, "Argument mismatch!")
	assert.Equal(t, int64(2), signature.NotSigned, "Argument mismatch!")
	assert.Equal(t, int64(0), signature.NotActive, "Argument mismatch!")
	assert.Equal(t, int64(0), signature.Proposed, "Argument mismatch!")
}

func TestValidatorsMissedBlocksAllInactive(t *testing.T) {
	t.Parallel()

	validator := &types.Validator{ConsensusAddressHex: "address"}
	state := NewState()

	state.AddBlock(&types.Block{Height: 1, Signatures: map[string]int32{}, Validators: map[string]bool{}})
	state.AddBlock(&types.Block{Height: 2, Signatures: map[string]int32{}, Validators: map[string]bool{}})
	state.AddBlock(&types.Block{Height: 3, Signatures: map[string]int32{}, Validators: map[string]bool{}})
	state.AddBlock(&types.Block{Height: 4, Signatures: map[string]int32{}, Validators: map[string]bool{}})
	state.AddBlock(&types.Block{Height: 5, Signatures: map[string]int32{}, Validators: map[string]bool{}})

	signature, err := state.GetValidatorMissedBlocks(validator, 5)

	require.NoError(t, err, "Error should not be present!")
	assert.Equal(t, int64(0), signature.Signed, "Argument mismatch!")
	assert.Equal(t, int64(0), signature.NoSignature, "Argument mismatch!")
	assert.Equal(t, int64(0), signature.NotSigned, "Argument mismatch!")
	assert.Equal(t, int64(5), signature.NotActive, "Argument mismatch!")
	assert.Equal(t, int64(0), signature.Proposed, "Argument mismatch!")
}

func TestValidatorsMissedBlocksSomeSkipped(t *testing.T) {
	t.Parallel()

	validator := &types.Validator{
		ConsensusAddressHex: "address",
		SigningInfo:         &types.SigningInfo{MissedBlocksCounter: 1},
	}
	state := NewState()

	state.AddBlock(&types.Block{Height: 1, Signatures: map[string]int32{}, Validators: map[string]bool{"address": true}})
	state.AddBlock(&types.Block{Height: 2, Signatures: map[string]int32{}, Validators: map[string]bool{"address": true}})
	state.AddBlock(&types.Block{Height: 3, Signatures: map[string]int32{}, Validators: map[string]bool{"address": true}})
	state.AddBlock(&types.Block{Height: 4, Signatures: map[string]int32{}, Validators: map[string]bool{"address": true}})
	state.AddBlock(&types.Block{Height: 5, Signatures: map[string]int32{}, Validators: map[string]bool{}})

	signature, err := state.GetValidatorMissedBlocks(validator, 5)

	require.NoError(t, err, "Error should not be present!")
	assert.Equal(t, int64(4), signature.Signed, "Argument mismatch!")
	assert.Equal(t, int64(0), signature.NoSignature, "Argument mismatch!")
	assert.Equal(t, int64(1), signature.NotSigned, "Argument mismatch!")
	assert.Equal(t, int64(1), signature.NotActive, "Argument mismatch!")
	assert.Equal(t, int64(0), signature.Proposed, "Argument mismatch!")
}
