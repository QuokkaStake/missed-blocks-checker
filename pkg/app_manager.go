package pkg

import (
	"fmt"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	dataPkg "main/pkg/data"
	databasePkg "main/pkg/database"
	"main/pkg/metrics"
	populatorsPkg "main/pkg/populators"
	reportersPkg "main/pkg/reporters"
	"main/pkg/reporters/discord"
	"main/pkg/reporters/telegram"
	snapshotPkg "main/pkg/snapshot"
	statePkg "main/pkg/state"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type AppManager struct {
	Logger             zerolog.Logger
	Config             *configPkg.ChainConfig
	Database           *databasePkg.Database
	DataManager        *dataPkg.Manager
	StateManager       *statePkg.Manager
	SnapshotManager    *snapshotPkg.Manager
	WebsocketManager   *tendermint.WebsocketManager
	MetricsManager     *metrics.Manager
	Populators         map[constants.PopulatorType]*populatorsPkg.Wrapper
	Reporters          []reportersPkg.Reporter
	IsPopulatingBlocks bool

	mutex         sync.Mutex
	snapshotMutex sync.Mutex
}

func NewAppManager(
	logger zerolog.Logger,
	config *configPkg.ChainConfig,
	version string,
	metricsManager *metrics.Manager,
	database *databasePkg.Database,
) *AppManager {
	managerLogger := logger.
		With().
		Str("chain", config.Name).
		Logger()

	dataManager := dataPkg.NewManager(managerLogger, config, metricsManager)
	snapshotManager := snapshotPkg.NewManager(managerLogger, config, metricsManager)
	stateManager := statePkg.NewManager(managerLogger, config, metricsManager, snapshotManager, database)
	websocketManager := tendermint.NewWebsocketManager(managerLogger, config, metricsManager)

	reporters := []reportersPkg.Reporter{
		telegram.NewReporter(config, version, managerLogger, stateManager, metricsManager, snapshotManager),
		discord.NewReporter(config, version, managerLogger, stateManager, metricsManager, snapshotManager),
	}

	populators := map[constants.PopulatorType]*populatorsPkg.Wrapper{
		constants.PopulatorSlashingParams: populatorsPkg.NewWrapper(
			populatorsPkg.NewSlashingParamsPopulator(config, dataManager, stateManager, metricsManager, managerLogger),
			config.Intervals.SlashingParams*time.Second,
			managerLogger,
		),
		constants.PopulatorTrimDatabase: populatorsPkg.NewWrapper(
			populatorsPkg.NewTrimDatabasePopulator(stateManager),
			config.Intervals.Trim*time.Second,
			managerLogger,
		),
	}

	return &AppManager{
		Logger:             managerLogger,
		Config:             config,
		DataManager:        dataManager,
		Database:           database,
		StateManager:       stateManager,
		SnapshotManager:    snapshotManager,
		WebsocketManager:   websocketManager,
		MetricsManager:     metricsManager,
		Reporters:          reporters,
		Populators:         populators,
		IsPopulatingBlocks: false,
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
		go reporter.Start()

		a.MetricsManager.LogReporterEnabled(a.Config.Name, reporter.Name(), reporter.Enabled())

		if reporter.Enabled() {
			a.Logger.Info().Str("name", string(reporter.Name())).Msg("Reporter is enabled")
		} else {
			a.Logger.Info().Str("name", string(reporter.Name())).Msg("Reporter is disabled")
		}
	}

	for _, populator := range a.Populators {
		go populator.Start()
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
			a.ProcessEvent(result)
		}
	}
}

func (a *AppManager) ProcessEvent(emittable types.WebsocketEmittable) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	block, ok := emittable.(*types.Block)
	if !ok {
		a.Logger.Warn().Msg("Event is not a block!")
		return
	}

	latestHeight := a.StateManager.GetLastBlockHeight()

	if latestHeight > block.Height {
		a.Logger.Warn().
			Int64("last_height", latestHeight).
			Int64("height", block.Height).
			Msg("Trying to generate a report for a block that was processed before")
		return
	}

	if err := a.UpdateValidators(block.Height - 1); err != nil {
		a.Logger.Error().
			Err(err).
			Msg("Error updating validators")
		return
	}

	validators, err := a.DataManager.GetActiveSetAtBlock(block.Height)
	if err != nil {
		a.Logger.Error().
			Err(err).
			Msg("Error updating historical validators")
		return
	}

	block.SetValidators(validators)

	a.Logger.Debug().Int64("height", block.Height).Msg("Got new block from Tendermint")
	if err := a.StateManager.AddBlock(block); err != nil {
		a.Logger.Error().
			Err(err).
			Msg("Error inserting new block")
	}

	a.ProcessSnapshot(block)
}

func (a *AppManager) ProcessSnapshot(block *types.Block) {
	a.snapshotMutex.Lock()
	defer a.snapshotMutex.Unlock()

	if a.SnapshotManager.HasNewerSnapshot() {
		if newerHeight := a.SnapshotManager.GetNewerHeight(); block.Height-newerHeight < a.Config.SnapshotsInterval {
			a.Logger.Debug().
				Int64("older_height", newerHeight).
				Int64("height", block.Height).
				Int64("diff", block.Height-newerHeight).
				Int64("snapshot_interval", a.Config.SnapshotsInterval).
				Msg("Trying to generate a snapshot between two blocks which are too close, skipping.")
			return
		}
	}

	totalBlocksCount := a.StateManager.GetBlocksCountSinceLatest(a.Config.StoreBlocks)
	a.Logger.Info().
		Int64("count", totalBlocksCount).
		Int64("height", block.Height).
		Msg("Added new Tendermint block into state")

	blocksCount := a.StateManager.GetBlocksCountSinceLatest(a.Config.BlocksWindow)

	neededBlocks := utils.MinInt64(a.Config.BlocksWindow, a.StateManager.GetLastBlockHeight()-a.Config.FirstBlock+1)
	hasEnoughBlocks := blocksCount >= neededBlocks

	if !hasEnoughBlocks {
		a.Logger.Info().
			Int64("blocks_count", blocksCount).
			Int64("expected", neededBlocks).
			Msg("Not enough data for producing a snapshot, skipping")
		return
	}

	snapshot, err := a.StateManager.GetSnapshot()
	if err != nil {
		a.Logger.Error().Err(err).Msg("Error generating snapshot")
		return
	}

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
		return
	}

	a.SnapshotManager.CommitNewSnapshot(block.Height, snapshot)
	if err := a.StateManager.SaveSnapshot(&snapshotPkg.Info{
		Height:   block.Height,
		Snapshot: snapshot,
	}); err != nil {
		a.Logger.Error().Err(err).Msg("Could not save latest snapshot to database")
	}

	olderHeight := a.SnapshotManager.GetOlderHeight()
	if olderHeight >= block.Height {
		a.Logger.Warn().
			Int64("older_height", olderHeight).
			Int64("height", block.Height).
			Msg("Trying to generate the snapshot for the older height, skipping.")
		return
	}

	a.Logger.Info().
		Int64("older_height", olderHeight).
		Int64("height", block.Height).
		Msg("Generating snapshot report")

	report, err := a.SnapshotManager.GetReport()
	if err != nil {
		a.Logger.Error().Err(err).Msg("Could not generate report")
		return
	}

	if report.Empty() {
		a.Logger.Info().Msg("Report is empty, no events to send")
		return
	}

	for _, event := range report.Events {
		a.Logger.Info().
			Str("event", fmt.Sprintf("%+v", event)).
			Msg("Report entries")
	}

	if err := a.StateManager.SaveReport(block.Height, report); err != nil {
		a.Logger.Error().
			Err(err).
			Msg("Error saving report to database")
	}

	for _, reporter := range a.Reporters {
		if reporter.Enabled() {
			if err := reporter.Send(report); err != nil {
				a.Logger.Error().
					Err(err).
					Str("name", string(reporter.Name())).
					Msg("Error sending report")
			}
		}
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

func (a *AppManager) PopulateInBackground() {
	// Start populating blocks in background
	go a.PopulateBlocks()

	// Setting timers
	go a.SyncBlocks()
}

func (a *AppManager) SyncBlocks() {
	if a.Config.Intervals.Blocks == 0 {
		a.Logger.Info().Msg("Blocks continuous population is disabled.")
		return
	}

	blocksTicker := time.NewTicker(a.Config.Intervals.Blocks * time.Second)
	for {
		select {
		case <-blocksTicker.C:
			a.PopulateBlocks()
		}
	}
}

func (a *AppManager) PopulateBlocks() {
	if a.IsPopulatingBlocks {
		a.Logger.Info().Msg("AppManager is populating blocks already, not populating again")
		return
	}

	a.IsPopulatingBlocks = true

	// Populating latest block
	a.Logger.Info().Msg("Populating blocks...")

	blockRaw, err := a.DataManager.GetBlock(0)
	if err != nil {
		a.Logger.Error().Err(err).Msg("Error querying for last block")
		a.IsPopulatingBlocks = false
		return
	}

	block, err := blockRaw.Result.Block.ToBlock()
	if err != nil {
		a.Logger.Warn().Msg("Error parsing block")
		a.IsPopulatingBlocks = false
		return
	}

	lastStateHeight := a.StateManager.GetLastBlockHeight()
	if lastStateHeight > block.Height {
		a.Logger.Info().
			Int64("last_height", lastStateHeight).
			Int64("height", block.Height).
			Msg("Got older block when populating latest height, not proceeding further.")
		a.IsPopulatingBlocks = false
		return
	}

	a.Logger.Info().
		Int64("height", block.Height).
		Time("time", block.Time).
		Msg("Last block height")

	validators, err := a.DataManager.GetActiveSetAtBlock(block.Height)
	if err != nil {
		a.Logger.Error().
			Err(err).
			Msg("Error updating historical validators")
		a.IsPopulatingBlocks = false
		return
	}

	block.SetValidators(validators)

	if err := a.StateManager.AddBlock(block); err != nil {
		a.Logger.Error().
			Err(err).
			Msg("Error inserting last block")
		a.IsPopulatingBlocks = false
		return
	}

	// Populating blocks
	if a.StateManager.GetLastBlockHeight() == 0 {
		a.Logger.Warn().Msg("Latest block is not set, cannot populate blocks.")
		a.IsPopulatingBlocks = false
		return
	}

	missingBlocks := a.StateManager.GetMissingBlocksSinceLatest(a.Config.StoreBlocks, a.Config.FirstBlock)
	if len(missingBlocks) == 0 {
		a.Logger.Info().
			Int64("count", a.Config.StoreBlocks).
			Msg("Got enough blocks for populating")
		a.IsPopulatingBlocks = false
		return
	}

	blocksChunks := utils.SplitIntoChunks(missingBlocks, a.Config.Pagination.BlocksSearch)

	for _, chunk := range blocksChunks {
		count := a.StateManager.GetBlocksCountSinceLatest(a.Config.StoreBlocks)

		a.Logger.Info().
			Int64("count", count).
			Int64("required", a.Config.StoreBlocks).
			Int("needed_blocks", len(chunk)).
			Ints64("blocks", chunk).
			Msg("Fetching more blocks...")

		blocks, allValidators, errs := a.DataManager.GetBlocksAndValidatorsAtHeights(chunk)

		if len(errs) > 0 {
			a.Logger.Error().Errs("errors", errs).Msg("Error querying for blocks")
		}

		for _, height := range chunk {
			blockRaw, found := blocks[height]
			if !found {
				a.Logger.Error().
					Int64("height", height).
					Msg("Could not find block at height")
				continue
			}

			validators, found := allValidators[height]
			if !found {
				a.Logger.Error().
					Int64("height", height).
					Msg("Could not find historical validators at height")
				continue
			}

			block, err := blockRaw.Result.Block.ToBlock()
			if err != nil {
				a.Logger.Error().
					Err(err).
					Msg("Error getting older block")
				continue
			}

			block.SetValidators(validators)

			a.mutex.Lock()

			if err := a.StateManager.AddBlock(block); err != nil {
				a.Logger.Error().
					Err(err).
					Msg("Error inserting older block")
				a.IsPopulatingBlocks = false
				a.mutex.Unlock()
				return
			}

			a.mutex.Unlock()
		}

		a.Logger.Debug().Int("len", len(blocks)).Msg("Inserted all blocks")
	}

	a.IsPopulatingBlocks = false

	latestHeight := a.StateManager.GetLastBlockHeight()

	if latestHeight > block.Height {
		a.Logger.Warn().
			Int64("last_height", latestHeight).
			Int64("height", block.Height).
			Msg("Trying to generate a report for a block that was processed before")
		return
	}

	if err := a.UpdateValidators(latestHeight - 1); err != nil {
		a.Logger.Error().
			Err(err).
			Msg("Error updating validators")
		return
	}

	a.ProcessSnapshot(block)
}
