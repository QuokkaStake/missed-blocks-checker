package snapshot

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"testing"

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
	assert.Nil(t, err, "Error should not be present!")
	assert.Len(t, report.Entries, 1, "Report should have 1 entry!")
	assert.Equal(t, report.Entries[0].Type(), constants.EventValidatorCreated, 1, "ReportEntry type mismatch!")
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
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 50},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	assert.Nil(t, err, "Error should not be present!")
	assert.NotEmpty(t, report.Entries, "Report should not be empty!")
	assert.Len(t, report.Entries, 1, "Report should have exactly 1 entry!")
	assert.Equal(t, report.Entries[0].Type(), constants.EventValidatorGroupChanged, 1, "ReportEntry type mismatch!")
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
	assert.Nil(t, err, "Error should not be present!")

	assert.NotEmpty(t, report.Entries, "Report should not be empty!")
	assert.Len(t, report.Entries, 1, "Report should have exactly 1 entry!")
	assert.Equal(t, report.Entries[0].Type(), constants.EventValidatorTombstoned, 1, "ReportEntry type mismatch!")
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
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			Validator:     &types.Validator{Jailed: true, Status: 1},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	assert.Nil(t, err, "Error should not be present!")

	assert.NotEmpty(t, report.Entries, "Report should not be empty!")
	assert.Len(t, report.Entries, 2, "Report should have exactly 2 entries!")

	entriesTypes := []constants.EventName{
		report.Entries[0].Type(),
		report.Entries[1].Type(),
	}

	assert.Contains(t, entriesTypes, constants.EventValidatorJailed, 1, "Expected to have jailed event!")
	assert.Contains(t, entriesTypes, constants.EventValidatorInactive, 1, "Expected to have inactive event!")
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
	assert.Nil(t, err, "Error should not be present!")

	assert.NotEmpty(t, report.Entries, "Report should not be empty!")
	assert.Len(t, report.Entries, 1, "Report should have exactly 1 entry!")
	assert.Equal(t, report.Entries[0].Type(), constants.EventValidatorUnjailed, 1, "ReportEntry type mismatch!")
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
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			Validator:     &types.Validator{Jailed: false, Status: 1},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	assert.Nil(t, err, "Error should not be present!")
	assert.NotEmpty(t, report.Entries, "Report should not be empty!")
	assert.Len(t, report.Entries, 1, "Report should have exactly 1 entry!")
	assert.Equal(t, report.Entries[0].Type(), constants.EventValidatorInactive, 1, "ReportEntry type mismatch!")
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
			Validator:     &types.Validator{Jailed: false, Status: 1},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	assert.Nil(t, err, "Error should not be present!")
	assert.NotEmpty(t, report.Entries, "Report should not be empty!")
	assert.Len(t, report.Entries, 1, "Report should have exactly 1 entry!")
	assert.Equal(t, report.Entries[0].Type(), constants.EventValidatorActive, 1, "ReportEntry type mismatch!")
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
	assert.Nil(t, err, "Error should not be present!")
	assert.Empty(t, report.Entries, "Report should be empty!")
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
	assert.Nil(t, err, "Error should not be present!")
	assert.Empty(t, report.Entries, "Report should be empty!")
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
	assert.NotEmpty(t, slice, "Slice should not be empty!")
	assert.Len(t, slice, 1, "Slice should have exactly 1 entry!")
	assert.Equal(t, slice[0].Validator.Moniker, "test", 1, "Validator name mismatch!")
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
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 150},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	assert.NotNil(t, err, "Error should be present!")
	assert.Nil(t, report, "Report should not be present!")
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
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 150},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator": {
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 0},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	assert.NotNil(t, err, "Error should be present!")
	assert.Nil(t, report, "Report should not be present!")
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
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 25},
		},
		"validator2": {
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 25},
		},
		"validator3": {
			Validator:     &types.Validator{Jailed: false, Status: 3, SigningInfo: &types.SigningInfo{Tombstoned: false}},
			SignatureInfo: types.SignatureInto{NotSigned: 25},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		"validator1": {
			Validator:     &types.Validator{Jailed: true, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 25},
		},
		"validator2": {
			Validator:     &types.Validator{Jailed: false, Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 75},
		},
		"validator3": {
			Validator:     &types.Validator{Jailed: false, Status: 3, SigningInfo: &types.SigningInfo{Tombstoned: true}},
			SignatureInfo: types.SignatureInto{NotSigned: 25},
		},
	}}

	report, err := newerSnapshot.GetReport(olderSnapshot, config)
	assert.Nil(t, err, "Error should not be present!")
	assert.NotNil(t, report, "Report should be present!")
	assert.Len(t, report.Entries, 3, "Slice should have exactly 3 entries!")
	assert.Equal(t, report.Entries[0].Type(), constants.EventValidatorTombstoned, "ReportEntry type mismatch!")
	assert.Equal(t, report.Entries[1].Type(), constants.EventValidatorJailed, "ReportEntry type mismatch!")
	assert.Equal(t, report.Entries[2].Type(), constants.EventValidatorGroupChanged, "ReportEntry type mismatch!")
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
			Validator:     &types.Validator{OperatorAddress: "validator1", Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 25},
		},
		"validator2": {
			Validator:     &types.Validator{OperatorAddress: "validator2", Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 75},
		},
		"validator3": {
			Validator:     &types.Validator{OperatorAddress: "validator3", Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 125},
		},
		"validator4": {
			Validator:     &types.Validator{OperatorAddress: "validator4", Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 75},
		},
	}}
	newerSnapshot := Snapshot{Entries: map[string]Entry{
		// skipping blocks: 25 -> 75
		"validator1": {
			Validator:     &types.Validator{OperatorAddress: "validator1", Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 75},
		},
		// skipping blocks: 75 -> 125
		"validator2": {
			Validator:     &types.Validator{OperatorAddress: "validator2", Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 125},
		},
		// recovering: 125 -> 75
		"validator3": {
			Validator:     &types.Validator{OperatorAddress: "validator3", Status: 3},
			SignatureInfo: types.SignatureInto{NotSigned: 75},
		},
		// recovering: 75 -> 25
		"validator4": {
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
	assert.Nil(t, err, "Error should not be present!")
	assert.NotNil(t, report, "Report should be present!")
	assert.Len(t, report.Entries, 4, "Slice should have exactly 4 entries!")
	assert.Equal(t, report.Entries[0].GetValidator().OperatorAddress, "validator2", "Wrong validator!")
	assert.Equal(t, report.Entries[1].GetValidator().OperatorAddress, "validator1", "Wrong validator!")
	assert.Equal(t, report.Entries[2].GetValidator().OperatorAddress, "validator3", "Wrong validator!")
	assert.Equal(t, report.Entries[3].GetValidator().OperatorAddress, "validator4", "Wrong validator!")
}
