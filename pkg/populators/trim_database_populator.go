package populators

import (
	"main/pkg/state"
)

type TrimDatabasePopulator struct {
	StateManager *state.Manager
}

func NewTrimDatabasePopulator(
	stateManager *state.Manager,
) *TrimDatabasePopulator {
	return &TrimDatabasePopulator{
		StateManager: stateManager,
	}
}
func (p *TrimDatabasePopulator) Populate() error {
	return p.StateManager.TrimBlocks()
}

func (p *TrimDatabasePopulator) Enabled() bool {
	return true
}

func (p *TrimDatabasePopulator) Name() string {
	return "trim-database-populator"
}
