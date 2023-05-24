package snapshot

import (
	configPkg "main/pkg/config"
	reportPkg "main/pkg/report"

	"github.com/rs/zerolog"
)

type Manager struct {
	logger zerolog.Logger
	config *configPkg.ChainConfig

	olderSnapshot *Info
	newerSnapshot *Info
}

func NewManager(logger zerolog.Logger, config *configPkg.ChainConfig) *Manager {
	return &Manager{
		logger: logger.With().Str("component", "state_manager").Logger(),
		config: config,
	}
}

func (m *Manager) CommitNewSnapshot(height int64, snapshot Snapshot) {
	m.olderSnapshot = m.newerSnapshot

	m.newerSnapshot = &Info{
		Snapshot: snapshot,
		Height:   height,
	}
}

func (m *Manager) HasNewerSnapshot() bool {
	return m.newerSnapshot != nil
}

func (m *Manager) GetOlderHeight() int64 {
	return m.olderSnapshot.Height
}

func (m *Manager) GetReport() *reportPkg.Report {
	return m.newerSnapshot.Snapshot.GetReport(m.olderSnapshot.Snapshot, m.config)
}
