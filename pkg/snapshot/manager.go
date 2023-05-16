package snapshot

import (
	configPkg "main/pkg/config"
	reportPkg "main/pkg/report"

	"github.com/rs/zerolog"
)

type Manager struct {
	logger zerolog.Logger
	config *configPkg.Config

	olderSnapshot       *Snapshot
	newerSnapshot       *Snapshot
	olderSnapshotHeight int64
	newerSnapshotHeight int64
}

func NewManager(logger zerolog.Logger, config *configPkg.Config) *Manager {
	return &Manager{
		logger: logger.With().Str("component", "state_manager").Logger(),
		config: config,
	}
}

func (m *Manager) CommitNewSnapshot(height int64, snapshot *Snapshot) {
	m.olderSnapshot = m.newerSnapshot
	m.olderSnapshotHeight = m.newerSnapshotHeight

	m.newerSnapshot = snapshot
	m.newerSnapshotHeight = height
}

func (m *Manager) HasOlderSnapshot() bool {
	return m.olderSnapshot != nil
}

func (m *Manager) GetOlderHeight() int64 {
	return m.olderSnapshotHeight
}

func (m *Manager) GetReport() *reportPkg.Report {
	return m.newerSnapshot.GetReport(m.olderSnapshot, m.config)
}
