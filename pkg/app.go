package pkg

import (
	"fmt"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	loggerPkg "main/pkg/logger"
	"main/pkg/metrics"
	reportersPkg "main/pkg/reporters"
	"main/pkg/reporters/telegram"
	snapshotPkg "main/pkg/snapshot"
	statePkg "main/pkg/state"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"
	"time"

	"github.com/rs/zerolog"
)

type App struct {
	Logger                zerolog.Logger
	Config                *configPkg.Config
	RPCManager            *tendermint.RPCManager
	StateManager          *statePkg.Manager
	SnapshotManager       *snapshotPkg.Manager
	WebsocketManager      *tendermint.WebsocketManager
	MetricsManager        *metrics.Manager
	Reporters             []reportersPkg.Reporter
	IsPopulatingBlocks    bool
	IsPopulatingActiveSet bool
	Version               string
}

func NewApp(configPath string, version string) *App {
	config, err := configPkg.GetConfig(configPath)
	if err != nil {
		loggerPkg.GetDefaultLogger().Fatal().Err(err).Msg("Could not load config")
	}
	config.SetDefaultMissedBlocksGroups()

	if err = config.Validate(); err != nil {
		loggerPkg.GetDefaultLogger().Fatal().Err(err).Msg("Provided config is invalid!")
	}

	logger := loggerPkg.GetLogger(config.LogConfig).
		With().
		Str("chain", config.ChainConfig.Name).
		Logger()

	metricsManager := metrics.NewManager(logger, config)
	rpcManager := tendermint.NewRPCManager(config.ChainConfig.RPCEndpoints, logger, metricsManager)
	snapshotManager := snapshotPkg.NewManager(logger, config)
	stateManager := statePkg.NewManager(logger, config, metricsManager, snapshotManager)
	websocketManager := tendermint.NewWebsocketManager(logger, config, metricsManager)

	reporters := []reportersPkg.Reporter{
		telegram.NewReporter(config, logger, stateManager),
	}

	return &App{
		Logger:                logger,
		Config:                config,
		RPCManager:            rpcManager,
		StateManager:          stateManager,
		SnapshotManager:       snapshotManager,
		WebsocketManager:      websocketManager,
		MetricsManager:        metricsManager,
		Reporters:             reporters,
		IsPopulatingBlocks:    false,
		IsPopulatingActiveSet: false,
		Version:               version,
	}
}

func (a *App) Start() {
	a.MetricsManager.LogAppVersion(a.Version)

	a.StateManager.Init()

	go a.MetricsManager.Start()

	for _, reporter := range a.Reporters {
		reporter.Init()

		a.MetricsManager.LogReporterEnabled(reporter.Name(), reporter.Enabled())

		if reporter.Enabled() {
			a.Logger.Debug().Str("name", reporter.Name()).Msg("Reporter is enabled")
		} else {
			a.Logger.Debug().Str("name", reporter.Name()).Msg("Reporter is disabled")
		}
	}

	go a.ListenForEvents()
	go a.PopulateInBackground()

	select {}
}

func (a *App) ListenForEvents() {
	a.WebsocketManager.Listen()

	for {
		select {
		case result := <-a.WebsocketManager.Channel:
			block, ok := result.(*types.Block)
			if !ok {
				a.Logger.Warn().Msg("Event is not a block!")
				continue
			}

			if a.StateManager.HasBlockAtHeight(block.Height) {
				a.Logger.Info().
					Int64("height", block.Height).
					Msg("Already have block at this height, not generating report.")
				continue
			}

			if err := a.UpdateValidators(); err != nil {
				a.Logger.Error().
					Err(err).
					Msg("Error updating validators")
			}

			if err := a.AddLastActiveSet(block.Height); err != nil {
				a.Logger.Error().
					Err(err).
					Msg("Error updating historical validators")
			}

			a.Logger.Debug().Int64("height", block.Height).Msg("Got new block from Tendermint")
			if err := a.StateManager.AddBlock(block); err != nil {
				a.Logger.Error().
					Err(err).
					Msg("Error inserting new block")
			}

			blocksCount := a.StateManager.GetBlocksCountSinceLatest(a.Config.ChainConfig.StoreBlocks)
			a.Logger.Info().
				Int64("count", blocksCount).
				Int64("height", block.Height).
				Msg("Added new Tendermint block into state")

			historicalValidatorsCount := a.StateManager.GetActiveSetsCountSinceLatest(a.Config.ChainConfig.StoreBlocks)

			hasEnoughBlocks := blocksCount >= a.Config.ChainConfig.BlocksWindow
			hasEnoughHistoricalValidators := historicalValidatorsCount >= a.Config.ChainConfig.BlocksWindow

			if !hasEnoughBlocks || !hasEnoughHistoricalValidators {
				a.Logger.Info().
					Int64("blocks_count", blocksCount).
					Int64("historical_validators_count", historicalValidatorsCount).
					Int64("expected", a.Config.ChainConfig.BlocksWindow).
					Msg("Not enough data for producing a snapshot, skipping.")
				continue
			}

			snapshot := a.StateManager.GetSnapshot()

			for _, entry := range snapshot.Entries {
				a.Logger.Debug().
					Str("valoper", entry.Validator.OperatorAddress).
					Str("moniker", entry.Validator.Moniker).
					Int64("signed", entry.SignatureInfo.Signed).
					Int64("not_signed", entry.SignatureInfo.NotSigned).
					Int64("no_signature", entry.SignatureInfo.NoSignature).
					Int64("not_active", entry.SignatureInfo.NotActive).
					Int64("proposed", entry.SignatureInfo.Proposed).
					Msg("Validator signing info")
			}

			if !a.SnapshotManager.HasNewerSnapshot() {
				a.Logger.Info().Msg("No older snapshot present, cannot generate report")
				a.SnapshotManager.CommitNewSnapshot(block.Height, snapshot)
				continue
			}

			a.SnapshotManager.CommitNewSnapshot(block.Height, snapshot)
			if err := a.StateManager.SaveSnapshot(&snapshotPkg.Info{
				Height:   block.Height,
				Snapshot: snapshot,
			}); err != nil {
				a.Logger.Error().Err(err).Msg("Could not save latest snapshot to database")
			}

			olderHeight := a.SnapshotManager.GetOlderHeight()
			a.Logger.Info().
				Int64("older_height", olderHeight).
				Int64("height", block.Height).
				Msg("Generating snapshot report")

			report := a.SnapshotManager.GetReport()

			if report.Empty() {
				a.Logger.Info().Msg("Report is empty, no events to send.")
				continue
			}

			a.MetricsManager.LogReport(report)

			for _, entry := range report.Entries {
				a.Logger.Info().
					Str("entry", fmt.Sprintf("%+v", entry)).
					Msg("Report entry")
			}

			for _, reporter := range a.Reporters {
				if err := reporter.Send(report); err != nil {
					a.Logger.Error().
						Err(err).
						Str("name", reporter.Name()).
						Msg("Error sending report")
				}
			}
		}
	}
}

func (a *App) UpdateValidators() error {
	validators, err := a.RPCManager.GetValidators()
	if err != nil {
		return err
	}

	a.StateManager.SetValidators(validators.ToMap())
	return nil
}

func (a *App) AddLastActiveSet(height int64) error {
	validators, err := a.RPCManager.GetActiveSetAtBlock(height)
	if err != nil {
		return err
	}

	return a.StateManager.AddActiveSet(height, validators)
}

func (a *App) PopulateInBackground() {
	a.PopulateLatestBlock()

	go a.PopulateBlocks()
	go a.PopulateActiveSet()

	blocksTicker := time.NewTicker(60 * time.Second)
	activeSetTicker := time.NewTicker(60 * time.Second)
	latestBlockTimer := time.NewTicker(120 * time.Second)
	trimTimer := time.NewTicker(300 * time.Second)
	quit := make(chan struct{})

	for {
		select {
		case <-blocksTicker.C:
			a.PopulateBlocks()
		case <-activeSetTicker.C:
			a.PopulateActiveSet()
		case <-latestBlockTimer.C:
			a.PopulateLatestBlock()
		case <-trimTimer.C:
			{
				if err := a.StateManager.TrimBlocks(); err != nil {
					a.Logger.Error().Err(err).Msg("Error trimming blocks")
				}
				if err := a.StateManager.TrimHistoricalValidators(); err != nil {
					a.Logger.Error().Err(err).Msg("Error trimming historical validators")
				}
			}
		case <-quit:
			blocksTicker.Stop()
			activeSetTicker.Stop()
			return
		}
	}
}

func (a *App) PopulateLatestBlock() {
	blockRaw, err := a.RPCManager.GetBlock(0)
	if err != nil {
		a.Logger.Error().Err(err).Msg("Error querying for last block")
		return
	}

	blockParsed, err := blockRaw.Result.Block.ToBlock()
	if err != nil {
		a.Logger.Error().Err(err).Msg("Error fetching last block")
		return
	}

	a.Logger.Info().
		Int64("height", blockParsed.Height).
		Msg("Last block height")

	if err := a.StateManager.AddBlock(blockParsed); err != nil {
		a.Logger.Error().
			Err(err).
			Msg("Error inserting last block")
		return
	}
}

func (a *App) PopulateBlocks() {
	if a.IsPopulatingBlocks {
		a.Logger.Debug().Msg("App is populating blocks already, not populating again")
		return
	}

	a.Logger.Debug().Msg("Populating blocks...")

	a.IsPopulatingBlocks = true

	missingBlocks := a.StateManager.GetMissingBlocksSinceLatest(a.Config.ChainConfig.StoreBlocks)
	if len(missingBlocks) == 0 {
		a.Logger.Info().
			Int64("count", a.Config.ChainConfig.StoreBlocks).
			Msg("Got enough blocks for populating")
		a.IsPopulatingBlocks = false
		return
	}

	blocksChunks := utils.SplitIntoChunks(missingBlocks, int(constants.BlockSearchPagination))

	for _, chunk := range blocksChunks {
		count := a.StateManager.GetBlocksCountSinceLatest(a.Config.ChainConfig.StoreBlocks)

		a.Logger.Info().
			Int64("count", count).
			Int64("required", a.Config.ChainConfig.StoreBlocks).
			Int64("needed_blocks", constants.BlockSearchPagination).
			Ints64("blocks", chunk).
			Msg("Fetching more blocks...")

		blocks, errs := a.RPCManager.GetBlocksAtHeights(chunk)

		if len(errs) > 0 {
			a.Logger.Error().Errs("errors", errs).Msg("Error querying for blocks")
			a.IsPopulatingBlocks = false
			return
		}

		for _, blockRaw := range blocks {
			block, err := blockRaw.Result.Block.ToBlock()
			if err != nil {
				a.Logger.Error().
					Err(err).
					Msg("Error getting older block")
				a.IsPopulatingBlocks = false
				return
			}

			if err := a.StateManager.AddBlock(block); err != nil {
				a.Logger.Error().
					Err(err).
					Msg("Error inserting older block")
				a.IsPopulatingBlocks = false
				return
			}
		}
	}

	a.IsPopulatingBlocks = false
}

func (a *App) PopulateActiveSet() {
	if a.IsPopulatingActiveSet {
		a.Logger.Debug().Msg("App is populating active set already, not populating again")
		return
	}

	a.Logger.Debug().Msg("Populating active set...")

	a.IsPopulatingActiveSet = true

	missing := a.StateManager.GetMissingHistoricalValidatorsSinceLatest(a.Config.ChainConfig.StoreBlocks)
	if len(missing) == 0 {
		a.Logger.Info().
			Int64("count", a.Config.ChainConfig.StoreBlocks).
			Msg("Got enough historical validators for populating")
		a.IsPopulatingActiveSet = false
		return
	}

	chunks := utils.SplitIntoChunks(missing, int(constants.ActiveSetsBulkQueryCount))
	for _, chunk := range chunks {
		count := a.StateManager.GetBlocksCountSinceLatest(a.Config.ChainConfig.StoreBlocks)

		a.Logger.Info().
			Int64("count", count).
			Ints64("blocks_to_fetch", chunk).
			Int64("required", a.Config.ChainConfig.StoreBlocks).
			Msg("Not enough historical validators, fetching more...")

		heightActiveSets, errs := a.RPCManager.GetActiveSetAtBlocks(chunk)
		if len(errs) > 0 {
			a.Logger.Error().
				Errs("errors", errs).
				Msg("Error querying for active set")
			a.IsPopulatingActiveSet = false
			return
		}

		for height, activeSet := range heightActiveSets {
			if err := a.StateManager.AddActiveSet(height, activeSet); err != nil {
				a.Logger.Error().
					Err(err).
					Msg("Error inserting active set")
				a.IsPopulatingActiveSet = false
				return
			}
		}
	}

	a.IsPopulatingActiveSet = false
}
