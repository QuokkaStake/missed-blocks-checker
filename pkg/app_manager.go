package pkg

import (
	"fmt"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	dataPkg "main/pkg/data"
	databasePkg "main/pkg/database"
	"main/pkg/metrics"
	reportersPkg "main/pkg/reporters"
	"main/pkg/reporters/discord"
	"main/pkg/reporters/telegram"
	snapshotPkg "main/pkg/snapshot"
	statePkg "main/pkg/state"
	templatesPkg "main/pkg/templates"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"
	"time"

	"github.com/rs/zerolog"
)

type AppManager struct {
	Logger                zerolog.Logger
	Config                *configPkg.ChainConfig
	RPCManager            *tendermint.RPCManager
	Database              *databasePkg.Database
	DataManager           *dataPkg.Manager
	StateManager          *statePkg.Manager
	SnapshotManager       *snapshotPkg.Manager
	WebsocketManager      *tendermint.WebsocketManager
	MetricsManager        *metrics.Manager
	Reporters             []reportersPkg.Reporter
	IsPopulatingBlocks    bool
	IsPopulatingActiveSet bool
}

func NewAppManager(
	logger zerolog.Logger,
	config *configPkg.ChainConfig,
	metricsManager *metrics.Manager,
	database *databasePkg.Database,
) *AppManager {
	managerLogger := logger.
		With().
		Str("chain", config.Name).
		Logger()

	rpcManager := tendermint.NewRPCManager(config, managerLogger, metricsManager)
	dataManager := dataPkg.NewManager(managerLogger, config, rpcManager)
	snapshotManager := snapshotPkg.NewManager(managerLogger, config, metricsManager)
	stateManager := statePkg.NewManager(managerLogger, config, metricsManager, snapshotManager, database)
	websocketManager := tendermint.NewWebsocketManager(managerLogger, config, metricsManager)

	templatesManager := templatesPkg.NewManager(logger)
	reporters := []reportersPkg.Reporter{
		telegram.NewReporter(config, managerLogger, stateManager, metricsManager, templatesManager),
		discord.NewReporter(config, managerLogger, stateManager, metricsManager, templatesManager),
	}

	return &AppManager{
		Logger:                managerLogger,
		Config:                config,
		RPCManager:            rpcManager,
		DataManager:           dataManager,
		Database:              database,
		StateManager:          stateManager,
		SnapshotManager:       snapshotManager,
		WebsocketManager:      websocketManager,
		MetricsManager:        metricsManager,
		Reporters:             reporters,
		IsPopulatingBlocks:    false,
		IsPopulatingActiveSet: false,
	}
}

func (a *AppManager) Start() {
	a.StateManager.Init()

	a.MetricsManager.LogSlashingParams(
		a.Config.Name,
		a.Config.BlocksWindow,
		a.Config.MinSignedPerWindow,
		a.Config.StoreBlocks,
	)
	a.MetricsManager.LogChainInfo(a.Config.Name, a.Config.GetName())

	for _, reporter := range a.Reporters {
		reporter.Init()

		a.MetricsManager.LogReporterEnabled(a.Config.Name, reporter.Name(), reporter.Enabled())

		if reporter.Enabled() {
			a.Logger.Info().Str("name", string(reporter.Name())).Msg("Reporter is enabled")
		} else {
			a.Logger.Info().Str("name", string(reporter.Name())).Msg("Reporter is disabled")
		}
	}

	go a.ListenForEvents()
	go a.PopulateInBackground()

	select {}
}

func (a *AppManager) ListenForEvents() {
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

			if err := a.UpdateValidators(block.Height - 1); err != nil {
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

			totalBlocksCount := a.StateManager.GetBlocksCountSinceLatest(a.Config.StoreBlocks)
			a.Logger.Info().
				Int64("count", totalBlocksCount).
				Int64("height", block.Height).
				Msg("Added new Tendermint block into state")

			blocksCount := a.StateManager.GetActiveSetsCountSinceLatest(a.Config.BlocksWindow)
			historicalValidatorsCount := a.StateManager.GetActiveSetsCountSinceLatest(a.Config.BlocksWindow)

			hasEnoughBlocks := blocksCount >= a.Config.BlocksWindow
			hasEnoughHistoricalValidators := historicalValidatorsCount >= a.Config.BlocksWindow

			if !hasEnoughBlocks || !hasEnoughHistoricalValidators {
				a.Logger.Info().
					Int64("blocks_count", blocksCount).
					Int64("historical_validators_count", historicalValidatorsCount).
					Int64("expected", a.Config.BlocksWindow).
					Msg("Not enough data for producing a snapshot, skipping.")
				continue
			}

			snapshot := a.StateManager.GetSnapshot()

			for _, entry := range snapshot.Entries {
				a.Logger.Trace().
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

			report, err := a.SnapshotManager.GetReport()
			if err != nil {
				a.Logger.Error().Err(err).Msg("Could not generate report")
				continue
			}

			if report.Empty() {
				a.Logger.Info().Msg("Report is empty, no events to send.")
				continue
			}

			for _, entry := range report.Entries {
				a.Logger.Info().
					Str("entry", fmt.Sprintf("%+v", entry)).
					Msg("Report entry")
			}

			for _, reporter := range a.Reporters {
				if err := reporter.Send(report); err != nil {
					a.Logger.Error().
						Err(err).
						Str("name", string(reporter.Name())).
						Msg("Error sending report")
				}
			}
		}
	}
}

func (a *AppManager) PopulateSlashingParams() {
	if !a.Config.QuerySlashingParams.Bool {
		return
	}

	if params, err := a.RPCManager.GetSlashingParams(a.StateManager.GetLastBlockHeight() - 1); err != nil {
		a.Logger.Warn().
			Err(err).
			Msg("Error updating slashing params")

		return
	} else {
		minSignedPerWindow, err := params.Params.MinSignedPerWindow.Float64()
		if err != nil {
			a.Logger.Warn().
				Err(err).
				Msg("Got malformed slashing params from node")
			return
		}

		a.Config.BlocksWindow = params.Params.SignedBlocksWindow
		a.Config.MinSignedPerWindow = minSignedPerWindow

		a.Logger.Info().
			Int64("blocks_window", a.Config.BlocksWindow).
			Float64("min_signed_per_window", a.Config.MinSignedPerWindow).
			Msg("Got slashing params")

		a.MetricsManager.LogSlashingParams(
			a.Config.Name,
			a.Config.BlocksWindow,
			a.Config.MinSignedPerWindow,
			a.Config.StoreBlocks,
		)
		a.Config.RecalculateMissedBlocksGroups()
	}
}

func (a *AppManager) UpdateValidators(height int64) error {
	validators, err := a.DataManager.GetValidators(height)
	if err != nil {
		return err
	}

	a.StateManager.SetValidators(validators.ToMap())
	return nil
}

func (a *AppManager) AddLastActiveSet(height int64) error {
	validators, err := a.RPCManager.GetActiveSetAtBlock(height)
	if err != nil {
		return err
	}

	return a.StateManager.AddActiveSet(height, validators)
}

func (a *AppManager) PopulateInBackground() {
	a.PopulateLatestBlock()
	a.PopulateSlashingParams()

	go a.PopulateBlocks()
	go a.PopulateActiveSet()

	blocksTicker := time.NewTicker(60 * time.Second)
	activeSetTicker := time.NewTicker(60 * time.Second)
	latestBlockTimer := time.NewTicker(120 * time.Second)
	trimTimer := time.NewTicker(300 * time.Second)
	slashingParamsTimer := time.NewTicker(300 * time.Second)

	quit := make(chan struct{})

	for {
		select {
		case <-blocksTicker.C:
			a.PopulateBlocks()
		case <-activeSetTicker.C:
			a.PopulateActiveSet()
		case <-latestBlockTimer.C:
			a.PopulateLatestBlock()
		case <-slashingParamsTimer.C:
			a.PopulateSlashingParams()
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

func (a *AppManager) PopulateLatestBlock() {
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

func (a *AppManager) PopulateBlocks() {
	if a.IsPopulatingBlocks {
		a.Logger.Debug().Msg("AppManager is populating blocks already, not populating again")
		return
	}

	a.Logger.Debug().Msg("Populating blocks...")

	a.IsPopulatingBlocks = true

	missingBlocks := a.StateManager.GetMissingBlocksSinceLatest(a.Config.StoreBlocks)
	if len(missingBlocks) == 0 {
		a.Logger.Info().
			Int64("count", a.Config.StoreBlocks).
			Msg("Got enough blocks for populating")
		a.IsPopulatingBlocks = false
		return
	}

	blocksChunks := utils.SplitIntoChunks(missingBlocks, int(constants.BlockSearchPagination))

	for _, chunk := range blocksChunks {
		count := a.StateManager.GetBlocksCountSinceLatest(a.Config.StoreBlocks)

		a.Logger.Info().
			Int64("count", count).
			Int64("required", a.Config.StoreBlocks).
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

func (a *AppManager) PopulateActiveSet() {
	if a.IsPopulatingActiveSet {
		a.Logger.Debug().Msg("AppManager is populating active set already, not populating again")
		return
	}

	a.Logger.Debug().Msg("Populating active set...")

	a.IsPopulatingActiveSet = true

	missing := a.StateManager.GetMissingHistoricalValidatorsSinceLatest(a.Config.StoreBlocks)
	if len(missing) == 0 {
		a.Logger.Info().
			Int64("count", a.Config.StoreBlocks).
			Msg("Got enough historical validators for populating")
		a.IsPopulatingActiveSet = false
		return
	}

	chunks := utils.SplitIntoChunks(missing, int(constants.ActiveSetsBulkQueryCount))
	for _, chunk := range chunks {
		count := a.StateManager.GetActiveSetsCountSinceLatest(a.Config.StoreBlocks)

		a.Logger.Info().
			Int64("count", count).
			Ints64("blocks_to_fetch", chunk).
			Int64("required", a.Config.StoreBlocks).
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
