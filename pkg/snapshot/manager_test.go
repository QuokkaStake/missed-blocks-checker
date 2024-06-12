package snapshot

import (
	configPkg "main/pkg/config"
	"main/pkg/logger"
	"main/pkg/metrics"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestManagerHasNewerSnapshot(t *testing.T) {
	t.Parallel()

	log := logger.GetDefaultLogger()
	config := &configPkg.ChainConfig{}
	manager := NewManager(*log, config, nil)
	assert.False(t, manager.HasNewerSnapshot(), "Should not have older snapshot!")
}

func TestManagerCommitNewSnapshot(t *testing.T) {
	t.Parallel()

	log := logger.GetDefaultLogger()
	config := &configPkg.ChainConfig{
		StoreBlocks: 10,
		Thresholds:  []float64{0, 100},
		EmojisStart: []string{"x"},
		EmojisEnd:   []string{"x"},
	}
	config.RecalculateMissedBlocksGroups()

	metricsManager := metrics.NewManager(*log, configPkg.MetricsConfig{Enabled: null.BoolFrom(true)})
	manager := NewManager(*log, config, metricsManager)

	manager.CommitNewSnapshot(10, Snapshot{
		Entries: map[string]types.Entry{
			"validator": {Validator: &types.Validator{}},
		},
	})
	manager.CommitNewSnapshot(20, Snapshot{
		Entries: map[string]types.Entry{
			"validator": {Validator: &types.Validator{}},
		},
	})

	assert.True(t, manager.HasNewerSnapshot(), "Should not have older snapshot!")
	assert.Equal(t, int64(10), manager.GetOlderHeight(), "Height mismatch!")
	assert.Equal(t, int64(20), manager.GetNewerHeight(), "Height mismatch!")

	report, err := manager.GetReport()
	require.NoError(t, err, "Error should not be presented!")
	assert.True(t, report.Empty(), "Report should be empty!")
}

func TestManagerGetNewerSnapshot(t *testing.T) {
	t.Parallel()

	log := logger.GetDefaultLogger()
	config := &configPkg.ChainConfig{
		StoreBlocks: 10,
		Thresholds:  []float64{0, 100},
		EmojisStart: []string{"x"},
		EmojisEnd:   []string{"x"},
	}
	config.RecalculateMissedBlocksGroups()

	metricsManager := metrics.NewManager(*log, configPkg.MetricsConfig{Enabled: null.BoolFrom(true)})
	manager := NewManager(*log, config, metricsManager)

	firstSnapshot, found := manager.GetNewerSnapshot()
	assert.False(t, found, "Snapshot should not be presented!")
	assert.Nil(t, firstSnapshot, "Snapshot should not be presented!")

	manager.CommitNewSnapshot(20, Snapshot{
		Entries: map[string]types.Entry{
			"validator": {Validator: &types.Validator{}},
		},
	})

	secondSnapshot, foundLater := manager.GetNewerSnapshot()
	assert.True(t, foundLater, "Snapshot should be presented!")
	assert.NotNil(t, secondSnapshot, "Snapshot should be presented!")
}
