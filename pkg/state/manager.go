package state

import (
	"github.com/rs/zerolog"
	configPkg "main/pkg/config"
	"main/pkg/types"
)

type Manager struct {
	Logger zerolog.Logger
	Config *configPkg.Config
	State  *State
}

func NewManager(logger *zerolog.Logger, config *configPkg.Config) *Manager {
	return &Manager{
		Logger: logger.With().Str("component", "state_manager").Logger(),
		Config: config,
		State:  NewState(),
	}
}

func (m *Manager) AddBlock(block *types.Block) {
	m.State.AddBlock(block)
}

func (m *Manager) GetBlocksCountSinceLatest(expected int64) int64 {
	return m.State.GetBlocksCountSinceLatest(expected)
}
