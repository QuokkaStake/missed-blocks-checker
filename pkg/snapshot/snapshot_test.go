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
	assert.Equal(t, report.Entries[0].Type(), constants.EventValidatorCreated, 1, "Entry type mismatch!")
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
	assert.Equal(t, report.Entries[0].Type(), constants.EventValidatorGroupChanged, 1, "Entry type mismatch!")
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
	assert.Equal(t, report.Entries[0].Type(), constants.EventValidatorTombstoned, 1, "Entry type mismatch!")
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
	assert.Equal(t, report.Entries[0].Type(), constants.EventValidatorUnjailed, 1, "Entry type mismatch!")
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
	assert.Equal(t, report.Entries[0].Type(), constants.EventValidatorInactive, 1, "Entry type mismatch!")
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
	assert.Equal(t, report.Entries[0].Type(), constants.EventValidatorActive, 1, "Entry type mismatch!")
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
	assert.Equal(t, report.Entries[0].Type(), constants.EventValidatorTombstoned, "Entry type mismatch!")
	assert.Equal(t, report.Entries[1].Type(), constants.EventValidatorJailed, "Entry type mismatch!")
	assert.Equal(t, report.Entries[2].Type(), constants.EventValidatorGroupChanged, "Entry type mismatch!")
}
