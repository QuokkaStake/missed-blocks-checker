package state

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	databasePkg "main/pkg/database"
	"main/pkg/metrics"
	snapshotPkg "main/pkg/snapshot"
	"main/pkg/types"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Manager struct {
	logger          zerolog.Logger
	config          *configPkg.ChainConfig
	metricsManager  *metrics.Manager
	snapshotManager *snapshotPkg.Manager
	state           *State
	database        *databasePkg.Database
	mutex           sync.Mutex
}

func NewManager(
	logger zerolog.Logger,
	chainConfig *configPkg.ChainConfig,
	metricsManager *metrics.Manager,
	snapshotManager *snapshotPkg.Manager,
	database *databasePkg.Database,
) *Manager {
	return &Manager{
		logger:          logger.With().Str("component", "state_manager").Logger(),
		config:          chainConfig,
		metricsManager:  metricsManager,
		snapshotManager: snapshotManager,
		state:           NewState(),
		database:        database,
	}
}

func (m *Manager) Init() {
	blocksStart := time.Now()

	blocks, err := m.database.GetAllBlocks(m.config.Name)
	if err != nil {
		m.logger.Fatal().Err(err).Msg("Could not get blocks from the database")
	}

	m.state.SetBlocks(blocks)
	m.logger.Info().
		Int("len", len(blocks)).
		Float64("duration", time.Since(blocksStart).Seconds()).
		Msg("Loaded older blocks from database")

	notifiersStart := time.Now()

	notifiers, err := m.database.GetAllNotifiers(m.config.Name)
	if err != nil {
		m.logger.Fatal().Err(err).Msg("Could not get notifiers from the database")
	}

	m.state.SetNotifiers(notifiers)
	m.logger.Info().
		Int("len", len(*notifiers)).
		Float64("duration", time.Since(notifiersStart).Seconds()).
		Msg("Loaded notifiers from database")

	activeSetStart := time.Now()

	activeSet, err := m.database.GetAllActiveSets(m.config.Name)
	if err != nil {
		m.logger.Fatal().Err(err).Msg("Could not get historical validators from the database")
	}

	m.state.SetActiveSet(activeSet)
	m.logger.Info().
		Int("len", len(activeSet)).
		Float64("duration", time.Since(activeSetStart).Seconds()).
		Msg("Loaded historical validators from database")

	snapshotStart := time.Now()

	snapshot, err := m.database.GetLastSnapshot(m.config.Name)
	if err != nil {
		m.logger.Error().Err(err).Msg("Could not get snapshot from the database")
	} else {
		m.logger.Info().
			Float64("duration", time.Since(snapshotStart).Seconds()).
			Msg("Loaded snapshot from database")
		m.snapshotManager.CommitNewSnapshot(snapshot.Height, snapshot.Snapshot)
	}
}

func (m *Manager) GetLastBlockHeight() int64 {
	return m.state.GetLastBlockHeight()
}

func (m *Manager) AddBlock(block *types.Block) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.state.AddBlock(block)

	if lastBlock := m.state.GetLastBlockHeight(); lastBlock == block.Height {
		m.metricsManager.LogLastHeight(m.config.Name, block.Height, block.Time)
	}

	if err := m.database.InsertBlock(m.config.Name, block); err != nil {
		return err
	}

	m.metricsManager.LogTotalBlocksAmount(m.config.Name, m.GetBlocksCountSinceLatest(m.config.StoreBlocks))

	return nil
}

func (m *Manager) AddActiveSet(height int64, activeSet map[string]bool) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.state.AddActiveSet(height, activeSet)

	if err := m.database.InsertActiveSet(m.config.Name, height, activeSet); err != nil {
		return err
	}
	m.metricsManager.LogTotalHistoricalValidatorsAmount(m.config.Name, m.GetActiveSetsCountSinceLatest(m.config.StoreBlocks))

	return nil
}

func (m *Manager) TrimBlocks() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	trimHeight := m.state.GetLastBlockHeight() - m.config.StoreBlocks
	m.logger.Info().
		Int64("height", m.state.GetLastBlockHeight()).
		Int64("trim_height", trimHeight).
		Msg("Need to trim blocks")

	m.state.TrimBlocksBefore(trimHeight)
	if err := m.database.TrimBlocksBefore(m.config.Name, trimHeight); err != nil {
		return err
	}

	return nil
}

func (m *Manager) TrimHistoricalValidators() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	trimHeight := m.state.GetLastBlockHeight() - m.config.StoreBlocks
	m.logger.Info().
		Int64("height", m.state.GetLastBlockHeight()).
		Int64("trim_height", trimHeight).
		Msg("Need to trim historical validators")

	m.state.TrimActiveSetsBefore(trimHeight)
	if err := m.database.TrimActiveSetsBefore(m.config.Name, trimHeight); err != nil {
		return err
	}

	return nil
}

func (m *Manager) HasBlockAtHeight(height int64) bool {
	return m.state.HasBlockAtHeight(height)
}

func (m *Manager) GetBlocksCountSinceLatest(expected int64) int64 {
	return m.state.GetBlocksCountSinceLatest(expected)
}

func (m *Manager) GetMissingBlocksSinceLatest(expected int64) []int64 {
	return m.state.GetMissingBlocksSinceLatest(expected)
}

func (m *Manager) GetActiveSetsCountSinceLatest(expected int64) int64 {
	return m.state.GetActiveSetsCountSinceLatest(expected)
}

func (m *Manager) GetMissingHistoricalValidatorsSinceLatest(expected int64) []int64 {
	return m.state.GetMissingActiveSetsSinceLatest(expected)
}

func (m *Manager) GetSnapshot() snapshotPkg.Snapshot {
	validators := m.state.GetValidators()
	entries := make(map[string]snapshotPkg.Entry, len(validators))

	for _, validator := range validators {
		entries[validator.OperatorAddress] = snapshotPkg.Entry{
			Validator:     validator,
			SignatureInfo: m.state.GetValidatorMissedBlocks(validator, m.config.BlocksWindow),
		}
	}

	return snapshotPkg.Snapshot{Entries: entries}
}

func (m *Manager) AddNotifier(
	operatorAddress string,
	reporter constants.ReporterName,
	userId string,
	userName string,
) bool {
	if added := m.state.AddNotifier(operatorAddress, reporter, userId, userName); !added {
		return false
	}

	err := m.database.InsertNotifier(m.config.Name, operatorAddress, reporter, userId, userName)
	return err == nil
}

func (m *Manager) RemoveNotifier(
	operatorAddress string,
	reporter constants.ReporterName,
	notifier string,
) bool {
	if removed := m.state.RemoveNotifier(operatorAddress, reporter, notifier); !removed {
		return false
	}

	err := m.database.RemoveNotifier(m.config.Name, operatorAddress, reporter, notifier)
	return err == nil
}

func (m *Manager) GetNotifiersForReporter(
	operatorAddress string,
	reporter constants.ReporterName,
) []*types.Notifier {
	return m.state.GetNotifiersForReporter(operatorAddress, reporter)
}

func (m *Manager) GetValidatorsForNotifier(
	reporter constants.ReporterName,
	notifier string,
) []string {
	return m.state.GetValidatorsForNotifier(reporter, notifier)
}

func (m *Manager) GetValidator(operatorAddress string) (*types.Validator, bool) {
	return m.state.GetValidator(operatorAddress)
}

func (m *Manager) GetValidators() types.ValidatorsMap {
	return m.state.GetValidators()
}

func (m *Manager) GetTimeTillJail(missingBlocks int64) time.Duration {
	return m.state.GetTimeTillJail(m.config, missingBlocks)
}

func (m *Manager) GetBlockTime() time.Duration {
	return m.state.GetBlockTime()
}

func (m *Manager) GetValidatorMissedBlocks(validator *types.Validator) types.SignatureInto {
	return m.state.GetValidatorMissedBlocks(validator, m.config.BlocksWindow)
}

func (m *Manager) SetValidators(validators types.ValidatorsMap) {
	m.state.SetValidators(validators)
}

func (m *Manager) SaveSnapshot(snapshot *snapshotPkg.Info) error {
	return m.database.SetSnapshot(m.config.Name, snapshot)
}

func (m *Manager) GetEarliestBlock() *types.Block {
	return m.state.GetEarliestBlock()
}
