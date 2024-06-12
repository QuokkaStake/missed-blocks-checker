package snapshot

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"testing"

	"github.com/stretchr/testify/require"

	"main/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestValidatorCreated(t *testing.T) {
	t.Parallel()

	olderSnapshot := Snapshot{Entries: map[string]Entry{}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {Validator: &types.Validator{}},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, nil)
	require.NoError(t, err)
	assert.Len(t, report.Events, 1, "Report should have 1 entry!")
	assert.Equal(t, constants.EventValidatorCreated, report.Events[0].Type())
}

func TestValidatorGroupChanged(t *testing.T) {
	t.Parallel()

	config := &configPkg.ChainConfig{
		MissedBlocksGroups: []*configPkg.MissedBlocksGroup{
			{Start: 0, End: 49},
			{Start: 50, End: 99},
		},
	}

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 50},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	require.NoError(t, err)
	assert.NotEmpty(t, report.Events)
	assert.Len(t, report.Events, 1)
	assert.Equal(t, constants.EventValidatorGroupChanged, report.Events[0].Type())
}

func TestValidatorGroupChangedAnomaly(t *testing.T) {
	t.Parallel()

	config := &configPkg.ChainConfig{
		MissedBlocksGroups: []*configPkg.MissedBlocksGroup{
			{Start: 0, End: 49},
			{Start: 50, End: 99},
			{Start: 100, End: 200},
		},
	}

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 125},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	require.NoError(t, err)
	assert.Empty(t, report.Events)
}

func TestValidatorTombstoned(t *testing.T) {
	t.Parallel()

	config := &configPkg.ChainConfig{
		MissedBlocksGroups: []*configPkg.MissedBlocksGroup{
			{Start: 0, End: 49},
			{Start: 50, End: 99},
		},
	}

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			Validator:     &types.Validator{SigningInfo: &types.SigningInfo{Tombstoned: false}},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			Validator:     &types.Validator{SigningInfo: &types.SigningInfo{Tombstoned: true}},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	require.NoError(t, err)

	assert.NotEmpty(t, report.Events)
	assert.Len(t, report.Events, 1)
	assert.Equal(t, constants.EventValidatorTombstoned, report.Events[0].Type())
}

func TestValidatorJailed(t *testing.T) {
	t.Parallel()

	config := &configPkg.ChainConfig{
		MissedBlocksGroups: []*configPkg.MissedBlocksGroup{
			{Start: 0, End: 49},
			{Start: 50, End: 99},
		},
	}

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      false,
			Validator:     &types.Validator{Jailed: true, Status: 1},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	require.NoError(t, err)

	assert.NotEmpty(t, report.Events)
	assert.Len(t, report.Events, 2)

	entriesTypes := []constants.EventName{
		report.Events[0].Type(),
		report.Events[1].Type(),
	}

	assert.Contains(t, entriesTypes, constants.EventValidatorJailed)
	assert.Contains(t, entriesTypes, constants.EventValidatorInactive)
}

func TestValidatorUnjailed(t *testing.T) {
	t.Parallel()

	config := &configPkg.ChainConfig{
		MissedBlocksGroups: []*configPkg.MissedBlocksGroup{
			{Start: 0, End: 49},
			{Start: 50, End: 99},
		},
	}

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			Validator:     &types.Validator{Jailed: true},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			Validator:     &types.Validator{Jailed: false},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	require.NoError(t, err)

	assert.NotEmpty(t, report.Events)
	assert.Len(t, report.Events, 1)
	assert.Equal(t, constants.EventValidatorUnjailed, report.Events[0].Type())
}

func TestValidatorJoinedSignatory(t *testing.T) {
	t.Parallel()

	config := &configPkg.ChainConfig{
		MissedBlocksGroups: []*configPkg.MissedBlocksGroup{
			{Start: 0, End: 49},
			{Start: 50, End: 99},
		},
	}

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      true,
			Validator:     &types.Validator{NeedsToSign: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      true,
			Validator:     &types.Validator{NeedsToSign: true, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	require.NoError(t, err)

	assert.NotEmpty(t, report.Events)
	assert.Len(t, report.Events, 1)
	assert.Equal(t, constants.EventValidatorJoinedSignatory, report.Events[0].Type())
}

func TestValidatorLeftSignatory(t *testing.T) {
	t.Parallel()

	config := &configPkg.ChainConfig{
		MissedBlocksGroups: []*configPkg.MissedBlocksGroup{
			{Start: 0, End: 49},
			{Start: 50, End: 99},
		},
	}

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      true,
			Validator:     &types.Validator{NeedsToSign: true, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      true,
			Validator:     &types.Validator{NeedsToSign: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	require.NoError(t, err)

	assert.NotEmpty(t, report.Events)
	assert.Len(t, report.Events, 1)
	assert.Equal(t, constants.EventValidatorLeftSignatory, report.Events[0].Type())
}

func TestValidatorInactive(t *testing.T) {
	t.Parallel()

	config := &configPkg.ChainConfig{
		MissedBlocksGroups: []*configPkg.MissedBlocksGroup{
			{Start: 0, End: 49},
			{Start: 50, End: 99},
		},
	}

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      false,
			Validator:     &types.Validator{Jailed: false, Status: 1},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	require.NoError(t, err)
	assert.NotEmpty(t, report.Events)
	assert.Len(t, report.Events, 1)
	assert.Equal(t, constants.EventValidatorInactive, report.Events[0].Type())
}

func TestValidatorChangedKey(t *testing.T) {
	t.Parallel()

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {Validator: &types.Validator{ConsensusAddressValcons: "key1"}},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {Validator: &types.Validator{ConsensusAddressValcons: "key2"}},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, nil)
	require.NoError(t, err)
	assert.Len(t, report.Events, 1)
	assert.Equal(t, constants.EventValidatorChangedKey, report.Events[0].Type())
}

func TestValidatorChangedMoniker(t *testing.T) {
	t.Parallel()

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {Validator: &types.Validator{Moniker: "moniker1"}},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {Validator: &types.Validator{Moniker: "moniker2"}},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, nil)
	require.NoError(t, err)
	assert.Len(t, report.Events, 1)
	assert.Equal(t, constants.EventValidatorChangedMoniker, report.Events[0].Type())
}

func TestValidatorChangedCommission(t *testing.T) {
	t.Parallel()

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {Validator: &types.Validator{Commission: 0.01}},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {Validator: &types.Validator{Commission: 0.02}},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, nil)
	require.NoError(t, err)
	assert.Len(t, report.Events, 1)
	assert.Equal(t, constants.EventValidatorChangedCommission, report.Events[0].Type())
}

func TestValidatorActive(t *testing.T) {
	t.Parallel()

	config := &configPkg.ChainConfig{
		MissedBlocksGroups: []*configPkg.MissedBlocksGroup{
			{Start: 0, End: 49},
			{Start: 50, End: 99},
		},
	}

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      false,
			Validator:     &types.Validator{Jailed: false, Status: 1},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	require.NoError(t, err)
	assert.NotEmpty(t, report.Events)
	assert.Len(t, report.Events, 1)
	assert.Equal(t, constants.EventValidatorActive, report.Events[0].Type())
}

func TestValidatorJailedAndChangedGroup(t *testing.T) {
	t.Parallel()

	config := &configPkg.ChainConfig{
		MissedBlocksGroups: []*configPkg.MissedBlocksGroup{
			{Start: 0, End: 49},
			{Start: 50, End: 99},
		},
	}

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			Validator:     &types.Validator{Jailed: true, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			Validator:     &types.Validator{Jailed: true, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 50},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	require.NoError(t, err)
	assert.Empty(t, report.Events)
}

func TestTombstonedAndNoPreviousSigningInfo(t *testing.T) {
	t.Parallel()

	config := &configPkg.ChainConfig{
		MissedBlocksGroups: []*configPkg.MissedBlocksGroup{
			{Start: 0, End: 49},
			{Start: 50, End: 99},
		},
	}

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			Validator:     &types.Validator{Jailed: true, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			Validator: &types.Validator{
				Jailed:      true,
				Status:      3,
				SigningInfo: &types.SigningInfo{Tombstoned: true},
			},
			SignatureInfo: types.SignatureInto{NotSigned: 50},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	require.NoError(t, err)
	assert.Empty(t, report.Events)
}

func TestToSlice(t *testing.T) {
	t.Parallel()

	entries := Entries{
		"validator": {
			Validator:     &types.Validator{Moniker: "test", Jailed: false, Status: 1},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}

	slice := entries.ToSlice()
	assert.NotEmpty(t, slice)
	assert.Len(t, slice, 1)
	assert.Equal(t, "test", slice[0].Validator.Moniker)
}

func TestNewMissedBlocksGroupNotPresent(t *testing.T) {
	t.Parallel()

	config := &configPkg.ChainConfig{
		MissedBlocksGroups: []*configPkg.MissedBlocksGroup{
			{Start: 0, End: 49},
			{Start: 50, End: 99},
		},
	}

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 150},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	require.Error(t, err)
	assert.Nil(t, report)
}

func TestOldMissedBlocksGroupNotPresent(t *testing.T) {
	t.Parallel()

	config := &configPkg.ChainConfig{
		MissedBlocksGroups: []*configPkg.MissedBlocksGroup{
			{Start: 0, End: 49},
			{Start: 50, End: 99},
		},
	}

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 150},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	require.Error(t, err)
	assert.Nil(t, report)
}

func TestSorting(t *testing.T) {
	t.Parallel()

	config := &configPkg.ChainConfig{
		MissedBlocksGroups: []*configPkg.MissedBlocksGroup{
			{Start: 0, End: 49},
			{Start: 50, End: 99},
		},
	}

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator1": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 25},
		},
		"validator2": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 25},
		},
		"validator3": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: false, Status: 3, SigningInfo: &types.SigningInfo{Tombstoned: false}},
			SignatureInfo: types.SignatureInto{NotSigned: 25},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator1": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: true, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 25},
		},
		"validator2": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 75},
		},
		"validator3": {
			IsActive:      true,
			Validator:     &types.Validator{Jailed: false, Status: 3, SigningInfo: &types.SigningInfo{Tombstoned: true}},
			SignatureInfo: types.SignatureInto{NotSigned: 25},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Len(t, report.Events, 3)
	assert.Equal(t, constants.EventValidatorTombstoned, report.Events[0].Type())
	assert.Equal(t, constants.EventValidatorJailed, report.Events[1].Type())
	assert.Equal(t, constants.EventValidatorGroupChanged, report.Events[2].Type())
}

func TestSortingMissedBlocksGroups(t *testing.T) {
	t.Parallel()

	config := &configPkg.ChainConfig{
		MissedBlocksGroups: []*configPkg.MissedBlocksGroup{
			{Start: 0, End: 49},
			{Start: 50, End: 99},
			{Start: 100, End: 149},
		},
	}

	olderSnapshot := Snapshot{Entries: map[string]Entry{
		"validator1": {
			IsActive:      true,
			Validator:     &types.Validator{OperatorAddress: "validator1", Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 25},
		},
		"validator2": {
			IsActive:      true,
			Validator:     &types.Validator{OperatorAddress: "validator2", Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 75},
		},
		"validator3": {
			IsActive:      true,
			Validator:     &types.Validator{OperatorAddress: "validator3", Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 125},
		},
		"validator4": {
			IsActive:      true,
			Validator:     &types.Validator{OperatorAddress: "validator4", Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 75},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		// skipping blocks: 25 -> 75
		"validator1": {
			IsActive:      true,
			Validator:     &types.Validator{OperatorAddress: "validator1", Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 75},
		},
		// skipping blocks: 75 -> 125
		"validator2": {
			IsActive:      true,
			Validator:     &types.Validator{OperatorAddress: "validator2", Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 125},
		},
		// recovering: 125 -> 75
		"validator3": {
			IsActive:      true,
			Validator:     &types.Validator{OperatorAddress: "validator3", Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 75},
		},
		// recovering: 75 -> 25
		"validator4": {
			IsActive:      true,
			Validator:     &types.Validator{OperatorAddress: "validator4", Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 25},
		},
	}}

	// expected order:
	// 1) validator2 skipping 75 -> 125
	// 2) validator1 skipping 25 -> 75
	// 3) validator3 recovering 125 -> 75
	// 4) validator4 recovering 75 -> 25

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Len(t, report.Events, 4, "Slice should have exactly 4 entries!")
	assert.Equal(t, "validator2", report.Events[0].GetValidator().OperatorAddress)
	assert.Equal(t, "validator1", report.Events[1].GetValidator().OperatorAddress)
	assert.Equal(t, "validator3", report.Events[2].GetValidator().OperatorAddress)
	assert.Equal(t, "validator4", report.Events[3].GetValidator().OperatorAddress)
}
