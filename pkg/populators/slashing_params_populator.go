package populators

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/data"
	"main/pkg/metrics"
	"main/pkg/state"

	"github.com/rs/zerolog"
)

type SlashingParamsPopulator struct {
	Config         *configPkg.ChainConfig
	DataManager    *data.Manager
	StateManager   *state.Manager
	MetricsManager *metrics.Manager
	Logger         zerolog.Logger
}

func NewSlashingParamsPopulator(
	config *configPkg.ChainConfig,
	dataManager *data.Manager,
	stateManager *state.Manager,
	metricsManager *metrics.Manager,
	logger zerolog.Logger,
) *SlashingParamsPopulator {
	return &SlashingParamsPopulator{
		Config:         config,
		DataManager:    dataManager,
		StateManager:   stateManager,
		MetricsManager: metricsManager,
		Logger: logger.With().
			Str("component", "slashing_params_populator").
			Logger(),
	}
}
func (p *SlashingParamsPopulator) Populate() error {
	params, err := p.DataManager.GetSlashingParams(p.StateManager.GetLastBlockHeight() - 1)
	if err != nil {
		p.Logger.Warn().
			Err(err).
			Msg("Error updating slashing params")

		return err
	}

	minSignedPerWindow := params.Params.MinSignedPerWindow.MustFloat64()

	p.Config.BlocksWindow = params.Params.SignedBlocksWindow
	p.Config.MinSignedPerWindow = minSignedPerWindow

	p.Logger.Info().
		Int64("blocks_window", p.Config.BlocksWindow).
		Float64("min_signed_per_window", p.Config.MinSignedPerWindow).
		Msg("Got slashing params")

	p.MetricsManager.LogSlashingParams(
		p.Config.Name,
		p.Config.BlocksWindow,
		p.Config.MinSignedPerWindow,
		p.Config.StoreBlocks,
	)
	p.Config.RecalculateMissedBlocksGroups()

	return nil
}

func (p *SlashingParamsPopulator) Enabled() bool {
	return true
}

func (p *SlashingParamsPopulator) Name() constants.PopulatorType {
	return constants.PopulatorSlashingParams
}
