package state

import (
	configPkg "main/pkg/config"
	"main/pkg/metrics"
	"main/pkg/snapshot"
	"main/pkg/types"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Manager struct {
	logger         zerolog.Logger
	config         *configPkg.Config
	metricsManager *metrics.Manager
	state          *State
	database       *Database
	mutex          sync.Mutex
}

func NewManager(logger zerolog.Logger, config *configPkg.Config, metricsManager *metrics.Manager) *Manager {
	return &Manager{
		logger:         logger.With().Str("component", "state_manager").Logger(),
		config:         config,
		metricsManager: metricsManager,
		state:          NewState(),
		database:       NewDatabase(logger, config),
	}
}

func (m *Manager) Init() {
	m.database.Init()

	blocksStart := time.Now()

	blocks, err := m.database.GetAllBlocks()
	if err != nil {
		m.logger.Fatal().Err(err).Msg("Could not get blocks from the database")
	}

	m.state.SetBlocks(blocks)
	m.logger.Info().
		Int("len", len(blocks)).
		Float64("duration", time.Since(blocksStart).Seconds()).
		Msg("Loaded older blocks from database")

	notifiersStart := time.Now()

	notifiers, err := m.database.GetAllNotifiers()
	if err != nil {
		m.logger.Fatal().Err(err).Msg("Could not get notifiers from the database")
	}

	m.state.SetNotifiers(notifiers)
	m.logger.Info().
		Int("len", len(*notifiers)).
		Float64("duration", time.Since(notifiersStart).Seconds()).
		Msg("Loaded notifiers from database")

	activeSetStart := time.Now()

	activeSet, err := m.database.GetAllActiveSets()
	if err != nil {
		m.logger.Fatal().Err(err).Msg("Could not get historical validators from the database")
	}

	m.state.SetActiveSet(activeSet)
	m.logger.Info().
		Int("len", len(activeSet)).
		Float64("duration", time.Since(activeSetStart).Seconds()).
		Msg("Loaded historical validators from database")
}

func (m *Manager) GetLatestBlock() int64 {
	return m.state.GetLatestBlock()
}

func (m *Manager) AddBlock(block *types.Block) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.state.AddBlock(block)

	if lastBlock := m.state.GetLatestBlock(); lastBlock == block.Height {
		m.metricsManager.LogLastHeight(block.Height, block.Time)
	}

	if err := m.database.InsertBlock(block); err != nil {
		return err
	}

	// newly added block, need to trim older blocks
	if m.state.GetLastBlockHeight() == block.Height {
		trimHeight := block.Height - m.config.ChainConfig.StoreBlocks
		m.logger.Debug().
			Int64("height", block.Height).
			Int64("trim_height", trimHeight).
			Msg("Need to trim blocks")

		m.state.TrimBlocksBefore(trimHeight)
		if err := m.database.TrimBlocksBefore(trimHeight); err != nil {
			return err
		}
	}

	m.metricsManager.LogTotalBlocksAmount(m.GetBlocksCountSinceLatest(m.config.ChainConfig.StoreBlocks))

	return nil
}

func (m *Manager) AddActiveSet(height int64, activeSet map[string]bool) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.state.AddActiveSet(height, activeSet)

	if err := m.database.InsertActiveSet(height, activeSet); err != nil {
		return err
	}

	// newly added block, need to trim older blocks
	if m.state.GetLastBlockHeight() == height {
		trimHeight := height - m.config.ChainConfig.StoreBlocks
		m.logger.Debug().
			Int64("height", height).
			Int64("trim_height", trimHeight).
			Msg("Need to trim active set")

		m.state.TrimActiveSetsBefore(trimHeight)
		if err := m.database.TrimActiveSetsBefore(trimHeight); err != nil {
			return err
		}
	}

	m.metricsManager.LogTotalHistoricalValidatorsAmount(m.GetActiveSetsCountSinceLatest(m.config.ChainConfig.StoreBlocks))

	return nil
}

func (m *Manager) HasBlockAtHeight(height int64) bool {
	return m.state.HasBlockAtHeight(height)
}

func (m *Manager) HasActiveSetAtHeight(height int64) bool {
	return m.state.HasActiveSetAtHeight(height)
}

func (m *Manager) GetBlocksCountSinceLatest(expected int64) int64 {
	return m.state.GetBlocksCountSinceLatest(expected)
}

func (m *Manager) GetActiveSetsCountSinceLatest(expected int64) int64 {
	return m.state.GetActiveSetsCountSinceLatest(expected)
}

func (m *Manager) GetSnapshot() *snapshot.Snapshot {
	validators := m.state.GetValidators()
	entries := make(map[string]snapshot.Entry, len(validators))

	for _, validator := range validators {
		entries[validator.OperatorAddress] = snapshot.Entry{
			Validator:     validator,
			SignatureInfo: m.state.GetValidatorMissedBlocks(validator, m.config.ChainConfig.BlocksWindow),
		}
	}

	return snapshot.NewSnapshot(entries)
}

func (m *Manager) AddNotifier(operatorAddress, reporter, notifier string) bool {
	if added := m.state.AddNotifier(operatorAddress, reporter, notifier); !added {
		return false
	}

	err := m.database.InsertNotifier(operatorAddress, reporter, notifier)
	return err == nil
}

func (m *Manager) RemoveNotifier(operatorAddress, reporter, notifier string) bool {
	if removed := m.state.RemoveNotifier(operatorAddress, reporter, notifier); !removed {
		return false
	}

	err := m.database.RemoveNotifier(operatorAddress, reporter, notifier)
	return err == nil
}

func (m *Manager) GetNotifiersForReporter(operatorAddress, reporter string) []string {
	return m.state.GetNotifiersForReporter(operatorAddress, reporter)
}

func (m *Manager) GetValidatorsForNotifier(reporter, notifier string) []string {
	return m.state.GetValidatorsForNotifier(reporter, notifier)
}

func (m *Manager) GetValidator(operatorAddress string) (*types.Validator, bool) {
	return m.state.GetValidator(operatorAddress)
}

func (m *Manager) GetValidators() types.ValidatorsMap {
	return m.state.GetValidators()
}

func (m *Manager) GetTimeTillJail(validator *types.Validator) (time.Duration, bool) {
	return m.state.GetTimeTillJail(validator, m.config)
}

func (m *Manager) GetValidatorMissedBlocks(validator *types.Validator) types.SignatureInto {
	return m.state.GetValidatorMissedBlocks(validator, m.config.ChainConfig.BlocksWindow)
}

func (m *Manager) SetValidators(validators types.ValidatorsMap) {
	m.state.SetValidators(validators)
}

func (m *Manager) GetEarliestBlock() *types.Block {
	return m.state.GetEarliestBlock()
}
