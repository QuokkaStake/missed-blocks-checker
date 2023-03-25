package state

import (
	"github.com/rs/zerolog"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/types"
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

	m.State.Blocks = blocks
	m.Logger.Info().Msg("Loaded older blocks from database")
}

func (m *Manager) AddBlock(block *types.Block) error {
	m.State.AddBlock(block)

	if err := m.Database.InsertBlock(block); err != nil {
		return err
	}

	// newly added block, need to trim older blocks
	if m.State.LastBlockHeight == block.Height {
		trimHeight := block.Height - constants.StoreBlocks
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
	entries := make(map[string]SnapshotEntry, len(m.State.Validators))

	for _, validator := range m.State.Validators {
		entries[validator.OperatorAddress] = SnapshotEntry{
			OperatorAddress: validator.OperatorAddress,
			Moniker:         validator.Moniker,
			Status:          validator.Status,
			Jailed:          validator.Jailed,
			SignatureInfo:   m.State.GetValidatorMissedBlocks(validator),
		}
	}

	return NewSnapshot(entries)
}
