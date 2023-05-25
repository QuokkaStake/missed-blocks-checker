package snapshot

import (
	"github.com/rs/zerolog"
	configPkg "main/pkg/config"
	"main/pkg/metrics"
	reportPkg "main/pkg/report"
)

type Manager struct {
	logger         zerolog.Logger
	config         *configPkg.ChainConfig
	metricsManager *metrics.Manager

	olderSnapshot *Info
	newerSnapshot *Info
}

func NewManager(
	logger zerolog.Logger,
	config *configPkg.ChainConfig,
	metricsManager *metrics.Manager,
) *Manager {
	return &Manager{
		logger:         logger.With().Str("component", "state_manager").Logger(),
		config:         config,
		metricsManager: metricsManager,
	}
}

func (m *Manager) CommitNewSnapshot(height int64, snapshot Snapshot) {
	m.olderSnapshot = m.newerSnapshot

	m.newerSnapshot = &Info{
		Snapshot: snapshot,
		Height:   height,
	}

	for _, entry := range snapshot.Entries {
		m.metricsManager.LogValidatorStats(m.config.Name, entry.Validator, entry.SignatureInfo)
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
