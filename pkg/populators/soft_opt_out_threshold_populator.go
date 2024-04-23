package populators

import (
	configPkg "main/pkg/config"
	"main/pkg/data"
	"main/pkg/metrics"
	"main/pkg/state"
	"strconv"

	"github.com/rs/zerolog"
)

type SoftOptOutThresholdPopulator struct {
	Config         *configPkg.ChainConfig
	DataManager    *data.Manager
	StateManager   *state.Manager
	MetricsManager *metrics.Manager
	Logger         zerolog.Logger
}

func NewSoftOptOutThresholdPopulator(
	config *configPkg.ChainConfig,
	dataManager *data.Manager,
	stateManager *state.Manager,
	metricsManager *metrics.Manager,
	logger zerolog.Logger,
) *SoftOptOutThresholdPopulator {
	return &SoftOptOutThresholdPopulator{
		Config:         config,
		DataManager:    dataManager,
		StateManager:   stateManager,
		MetricsManager: metricsManager,
		Logger: logger.With().
			Str("component", "soft_opt_out_threshold_populator").
			Logger(),
	}
}
func (p *SoftOptOutThresholdPopulator) Populate() error {
	params, err := p.DataManager.GetConsumerSoftOutOutThreshold(p.StateManager.GetLastBlockHeight() - 1)
	if err != nil {
		p.Logger.Warn().
			Err(err).
			Msg("Error updating soft out-out threshold")

		return err
	}

	thresholdAsString := params.Param.Value[1 : len(params.Param.Value)-1]
	threshold, err := strconv.ParseFloat(thresholdAsString, 64)
	if err != nil {
		p.Logger.Warn().
			Err(err).
			Msg("Error parsing soft out-out threshold")

		return err
	}

	p.Config.ConsumerSoftOptOut = threshold

	p.Logger.Info().
		Float64("threshold", threshold).
		Msg("Got soft out-out threshold")

	p.MetricsManager.LogConsumerSoftOutThreshold(p.Config.Name, threshold)
	return nil
}

func (p *SoftOptOutThresholdPopulator) Enabled() bool {
	return p.Config.IsConsumer.Bool
}

func (p *SoftOptOutThresholdPopulator) Name() string {
	return "soft-opt-out-threshold-populator"
}
