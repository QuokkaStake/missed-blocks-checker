package snapshot

import (
	configPkg "main/pkg/config"
	"main/pkg/logger"
	"main/pkg/metrics"
	"main/pkg/types"
	"testing"

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
		Entries: map[string]Entry{
			"validator": {Validator: &types.Validator{}},
		},
	})
	manager.CommitNewSnapshot(20, Snapshot{
		Entries: map[string]Entry{
			"validator": {Validator: &types.Validator{}},
		},
	})

	assert.True(t, manager.HasNewerSnapshot(), "Should not have older snapshot!")
	assert.Equal(t, manager.GetOlderHeight(), int64(10), "Height mismatch!")

	report, err := manager.GetReport()
	assert.Nil(t, err, "Error should not be presented!")
	assert.True(t, report.Empty(), "Report should be empty!")
}
