package state

import (
	configPkg "main/pkg/config"
	"main/pkg/types"
	"time"

	"github.com/rs/zerolog"
)

type Manager struct {
	Logger   zerolog.Logger
	Config   *configPkg.Config
	State    *State
	Database *Database
}

func NewManager(logger *zerolog.Logger, config *configPkg.Config) *Manager {
	return &Manager{
		Logger:   logger.With().Str("component", "state_manager").Logger(),
		Config:   config,
		State:    NewState(),
		Database: NewDatabase(logger, config),
	}
}

func (m *Manager) Init() {
	m.Database.Init()

	blocks, err := m.Database.GetAllBlocks()
	if err != nil {
		m.Logger.Fatal().Err(err).Msg("Could not get blocks from the database")
	}

	m.State.SetBlocks(blocks)
	m.Logger.Info().Msg("Loaded older blocks from database")

	notifiers, err := m.Database.GetAllNotifiers()
	if err != nil {
		m.Logger.Fatal().Err(err).Msg("Could not get notifiers from the database")
	}

	m.State.SetNotifiers(notifiers)
	m.Logger.Info().Msg("Loaded notifiers from database")
}

func (m *Manager) AddBlock(block *types.Block) error {
	m.State.AddBlock(block)

	if err := m.Database.InsertBlock(block); err != nil {
		return err
	}

	// newly added block, need to trim older blocks
	if m.State.GetLastBlockHeight() == block.Height {
		trimHeight := block.Height - m.Config.ChainConfig.StoreBlocks
		m.Logger.Debug().
			Int64("height", block.Height).
			Int64("trim_height", trimHeight).
			Msg("Need to trim blocks")

		m.State.TrimBlocksBefore(trimHeight)
		if err := m.Database.TrimBlocksBefore(trimHeight); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) GetBlocksCountSinceLatest(expected int64) int64 {
	return m.State.GetBlocksCountSinceLatest(expected)
}

func (m *Manager) GetSnapshot() *Snapshot {
	validators := m.State.GetValidators()
	entries := make(map[string]SnapshotEntry, len(validators))

	for _, validator := range validators {
		entries[validator.OperatorAddress] = SnapshotEntry{
			Validator:     validator,
			SignatureInfo: m.State.GetValidatorMissedBlocks(validator, m.Config.ChainConfig.BlocksWindow),
		}
	}

	return NewSnapshot(entries)
}

func (m *Manager) AddNotifier(operatorAddress, reporter, notifier string) bool {
	if added := m.State.AddNotifier(operatorAddress, reporter, notifier); !added {
		return false
	}

	err := m.Database.InsertNotifier(operatorAddress, reporter, notifier)
	return err == nil
}

func (m *Manager) RemoveNotifier(operatorAddress, reporter, notifier string) bool {
	if removed := m.State.RemoveNotifier(operatorAddress, reporter, notifier); !removed {
		return false
	}

	err := m.Database.RemoveNotifier(operatorAddress, reporter, notifier)
	return err == nil
}

func (m *Manager) GetValidator(operatorAddress string) (*types.Validator, bool) {
	return m.State.GetValidator(operatorAddress)
}

func (m *Manager) GetTimeTillJail(validator *types.Validator) (time.Duration, bool) {
	return m.State.GetTimeTillJail(validator, m.Config)
}
