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
			UserName:        "notifier",
			UserID:          "id",
		},
	})

	added := state.AddNotifier("address", constants.TelegramReporterName, "id", "notifier")

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
			UserName:        "notifier",
			UserID:          "id",
		},
	})

	added := state.AddNotifier("address", constants.TelegramReporterName, "id2", "newnotifier")
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
	assert.Equal(t, len(reporterNotifiers), 1, "Should have 1 notifier")
	assert.Equal(t, reporterNotifiers[0].UserName, "notifier2", "Should have 1 notifier")
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
	assert.Equal(t, validatorNotifiers[0], "address1", "Should have 1 notifier")
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
	assert.Equal(t, state.notifiers.Length(), 1, "New notifier should not be removed!")
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

func TestValidatorsMissedBlocksNoBlocks(t *testing.T) {
	t.Parallel()
	state := NewState()
	validator := &types.Validator{}
	signature := state.GetValidatorMissedBlocks(validator, 5)

	assert.Equal(t, signature.Signed, int64(0), "Argument mismatch!")
	assert.Equal(t, signature.NoSignature, int64(0), "Argument mismatch!")
	assert.Equal(t, signature.NotSigned, int64(0), "Argument mismatch!")
	assert.Equal(t, signature.NotActive, int64(0), "Argument mismatch!")
}

func TestValidatorsMissedBlocksAllSigned(t *testing.T) {
	t.Parallel()

	validator := &types.Validator{ConsensusAddressHex: "address"}
	state := NewState()
	state.AddActiveSet(1, types.HistoricalValidators{"address": true})
	state.AddActiveSet(2, types.HistoricalValidators{"address": true})
	state.AddActiveSet(3, types.HistoricalValidators{"address": true})
	state.AddActiveSet(4, types.HistoricalValidators{"address": true})
	state.AddActiveSet(5, types.HistoricalValidators{"address": true})

	state.AddBlock(&types.Block{Height: 1, Signatures: map[string]int32{"address": 3}, Proposer: "address"})
	state.AddBlock(&types.Block{Height: 2, Signatures: map[string]int32{"address": 3}})
	state.AddBlock(&types.Block{Height: 3, Signatures: map[string]int32{"address": 3}})
	state.AddBlock(&types.Block{Height: 4, Signatures: map[string]int32{"address": 3}})
	state.AddBlock(&types.Block{Height: 5, Signatures: map[string]int32{"address": 3}})

	signature := state.GetValidatorMissedBlocks(validator, 5)

	assert.Equal(t, signature.Signed, int64(5), "Argument mismatch!")
	assert.Equal(t, signature.NoSignature, int64(0), "Argument mismatch!")
	assert.Equal(t, signature.NotSigned, int64(0), "Argument mismatch!")
	assert.Equal(t, signature.NotActive, int64(0), "Argument mismatch!")
	assert.Equal(t, signature.Proposed, int64(1), "Argument mismatch!")
}

func TestValidatorsMissedBlocksAllMissed(t *testing.T) {
	t.Parallel()

	validator := &types.Validator{ConsensusAddressHex: "address"}
	state := NewState()
	state.AddActiveSet(1, types.HistoricalValidators{"address": true})
	state.AddActiveSet(2, types.HistoricalValidators{"address": true})
	state.AddActiveSet(3, types.HistoricalValidators{"address": true})
	state.AddActiveSet(4, types.HistoricalValidators{"address": true})
	state.AddActiveSet(5, types.HistoricalValidators{"address": true})

	state.AddBlock(&types.Block{Height: 1, Signatures: map[string]int32{}})
	state.AddBlock(&types.Block{Height: 2, Signatures: map[string]int32{}})
	state.AddBlock(&types.Block{Height: 3, Signatures: map[string]int32{}})
	state.AddBlock(&types.Block{Height: 4, Signatures: map[string]int32{"address": 1}})
	state.AddBlock(&types.Block{Height: 5, Signatures: map[string]int32{"address": 1}})

	signature := state.GetValidatorMissedBlocks(validator, 5)

	assert.Equal(t, signature.Signed, int64(0), "Argument mismatch!")
	assert.Equal(t, signature.NoSignature, int64(3), "Argument mismatch!")
	assert.Equal(t, signature.NotSigned, int64(2), "Argument mismatch!")
	assert.Equal(t, signature.NotActive, int64(0), "Argument mismatch!")
	assert.Equal(t, signature.Proposed, int64(0), "Argument mismatch!")
}

func TestValidatorsMissedBlocksAllInactive(t *testing.T) {
	t.Parallel()

	validator := &types.Validator{ConsensusAddressHex: "address"}
	state := NewState()
	state.AddActiveSet(1, types.HistoricalValidators{})
	state.AddActiveSet(2, types.HistoricalValidators{})
	state.AddActiveSet(3, types.HistoricalValidators{})
	state.AddActiveSet(4, types.HistoricalValidators{})
	state.AddActiveSet(5, types.HistoricalValidators{})

	state.AddBlock(&types.Block{Height: 1, Signatures: map[string]int32{}})
	state.AddBlock(&types.Block{Height: 2, Signatures: map[string]int32{}})
	state.AddBlock(&types.Block{Height: 3, Signatures: map[string]int32{}})
	state.AddBlock(&types.Block{Height: 4, Signatures: map[string]int32{}})
	state.AddBlock(&types.Block{Height: 5, Signatures: map[string]int32{}})

	signature := state.GetValidatorMissedBlocks(validator, 5)

	assert.Equal(t, signature.Signed, int64(0), "Argument mismatch!")
	assert.Equal(t, signature.NoSignature, int64(0), "Argument mismatch!")
	assert.Equal(t, signature.NotSigned, int64(0), "Argument mismatch!")
	assert.Equal(t, signature.NotActive, int64(5), "Argument mismatch!")
	assert.Equal(t, signature.Proposed, int64(0), "Argument mismatch!")
}

func TestValidatorsMissedBlocksSomeSkipped(t *testing.T) {
	t.Parallel()

	validator := &types.Validator{
		ConsensusAddressHex: "address",
		SigningInfo:         &types.SigningInfo{MissedBlocksCounter: 1},
	}
	state := NewState()
	state.AddActiveSet(1, types.HistoricalValidators{"address": true})
	state.AddActiveSet(2, types.HistoricalValidators{"address": true})
	state.AddActiveSet(3, types.HistoricalValidators{"address": true})
	state.AddActiveSet(4, types.HistoricalValidators{"address": true})
	state.AddActiveSet(5, types.HistoricalValidators{})

	state.AddBlock(&types.Block{Height: 1, Signatures: map[string]int32{}})
	state.AddBlock(&types.Block{Height: 2, Signatures: map[string]int32{}})
	state.AddBlock(&types.Block{Height: 3, Signatures: map[string]int32{}})
	state.AddBlock(&types.Block{Height: 4, Signatures: map[string]int32{}})
	state.AddBlock(&types.Block{Height: 5, Signatures: map[string]int32{}})

	signature := state.GetValidatorMissedBlocks(validator, 5)

	assert.Equal(t, signature.Signed, int64(4), "Argument mismatch!")
	assert.Equal(t, signature.NoSignature, int64(0), "Argument mismatch!")
	assert.Equal(t, signature.NotSigned, int64(1), "Argument mismatch!")
	assert.Equal(t, signature.NotActive, int64(1), "Argument mismatch!")
	assert.Equal(t, signature.Proposed, int64(0), "Argument mismatch!")
}
