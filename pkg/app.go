package pkg

import (
	"fmt"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/logger"
	reportersPkg "main/pkg/reporters"
	"main/pkg/reporters/telegram"
	statePkg "main/pkg/state"
	"main/pkg/tendermint"
	"main/pkg/types"
	"time"

	"github.com/rs/zerolog"
)

type App struct {
	Logger                zerolog.Logger
	Config                *configPkg.Config
	RPC                   *tendermint.RPC
	StateManager          *statePkg.Manager
	WebsocketManager      *tendermint.WebsocketManager
	Reporters             []reportersPkg.Reporter
	IsPopulatingBlocks    bool
	IsPopulatingActiveSet bool
}

func NewApp(configPath string) *App {
	config, err := configPkg.GetConfig(configPath)
	if err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not load config")
	}
	config.SetDefaultMissedBlocksGroups()

	if err = config.Validate(); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Provided config is invalid!")
	}

	log := logger.GetLogger(config.LogConfig).
		With().
		Str("chain", config.ChainConfig.Name).
		Logger()
	rpc := tendermint.NewRPC(config.ChainConfig.RPCEndpoints, log)
	stateManager := statePkg.NewManager(log, config)
	websocketManager := tendermint.NewWebsocketManager(log, config)

	reporters := []reportersPkg.Reporter{
		telegram.NewReporter(config, log, stateManager),
	}

	return &App{
		Logger:                log,
		Config:                config,
		RPC:                   rpc,
		StateManager:          stateManager,
		WebsocketManager:      websocketManager,
		Reporters:             reporters,
		IsPopulatingBlocks:    false,
		IsPopulatingActiveSet: false,
	}
}

func (a *App) Start() {
	a.StateManager.Init()

	for _, reporter := range a.Reporters {
		reporter.Init()

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

	var olderSnapshot *statePkg.Snapshot

	for {
		select {
		case result := <-a.WebsocketManager.Channel:
			block, ok := result.(*types.Block)
			if !ok {
				a.Logger.Warn().Msg("Event is not a block!")
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
					Msg("Error updating validators")
			}

			a.Logger.Debug().Int64("height", block.Height).Msg("Got new block from Tendermint")
			if err := a.StateManager.AddBlock(block); err != nil {
				a.Logger.Error().
					Err(err).
					Msg("Error inserting new block")
			}

			count := a.StateManager.GetBlocksCountSinceLatest(a.Config.ChainConfig.StoreBlocks)

			a.Logger.Info().
				Int64("count", count).
				Int64("height", block.Height).
				Msg("Added blocks into state")

			if !a.StateManager.IsPopulated() {
				a.Logger.Debug().
					Int64("count", count).
					Int64("expected", a.Config.ChainConfig.BlocksWindow).
					Msg("Not enough blocks for producing a snapshot, skipping.")
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

			if olderSnapshot == nil {
				a.Logger.Info().Msg("No older snapshot present, cannot generate report")
				olderSnapshot = snapshot
				continue
			}

			report := snapshot.GetReport(olderSnapshot, a.Config)
			olderSnapshot = snapshot

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
						Str("name", reporter.Name()).
						Msg("Error sending report")
				}
			}
		}
	}
}

func (a *App) UpdateValidators() error {
	validators, err := a.RPC.GetValidators()
	if err != nil {
		return err
	}

	a.StateManager.SetValidators(validators.ToMap())
	return nil
}

func (a *App) AddLastActiveSet(height int64) error {
	validators, err := a.RPC.GetActiveSetAtBlock(height)
	if err != nil {
		return err
	}

	return a.StateManager.AddActiveSet(height, validators)
}

func (a *App) PopulateInBackground() {
	a.PopulateBlocks()
	a.PopulateActiveSet()

	blocksTicker := time.NewTicker(60 * time.Second)
	activeSetTicker := time.NewTicker(60 * time.Second)
	quit := make(chan struct{})

	for {
		select {
		case <-blocksTicker.C:
			a.PopulateBlocks()
		case <-activeSetTicker.C:
			a.PopulateActiveSet()
		case <-quit:
			blocksTicker.Stop()
			activeSetTicker.Stop()
			return
		}
	}
}

func (a *App) PopulateBlocks() {
	if a.IsPopulatingBlocks {
		a.Logger.Debug().Msg("App is populating blocks already, not populating again")
		return
	}

	a.Logger.Debug().Msg("Populating blocks...")

	a.IsPopulatingBlocks = true

	block, err := a.RPC.GetLatestBlock()
	if err != nil {
		a.Logger.Error().Err(err).Msg("Error querying for last block")
		a.IsPopulatingBlocks = false
		return
	}

	blockParsed := block.Result.Block.ToBlock()

	a.Logger.Info().
		Int64("height", blockParsed.Height).
		Msg("Last chain block")

	if err := a.StateManager.AddBlock(blockParsed); err != nil {
		a.Logger.Error().
			Err(err).
			Msg("Error inserting last block")
		a.IsPopulatingBlocks = false
		return
	}

	startBlockToFetch := blockParsed.Height

	for {
		count := a.StateManager.GetBlocksCountSinceLatest(a.Config.ChainConfig.StoreBlocks)
		if count >= a.Config.ChainConfig.StoreBlocks {
			a.Logger.Info().
				Int64("count", count).
				Msg("Got enough blocks for populating")
			a.IsPopulatingBlocks = false
			break
		}

		a.Logger.Info().
			Int64("count", count).
			Int64("required", a.Config.ChainConfig.StoreBlocks).
			Msg("Not enough blocks, fetching more blocks...")

		blocks, err := a.RPC.GetBlocksFromTo(
			startBlockToFetch-constants.BlockSearchPagination,
			startBlockToFetch,
			constants.BlockSearchPagination,
		)
		if err != nil {
			a.Logger.Error().Err(err).Msg("Error querying for blocks search")
			a.IsPopulatingBlocks = false
			return
		}

		for _, block := range blocks.Result.Blocks {
			if err := a.StateManager.AddBlock(block.Block.ToBlock()); err != nil {
				a.Logger.Error().
					Err(err).
					Msg("Error inserting older block")
				a.IsPopulatingBlocks = false
				return
			}
		}

		startBlockToFetch -= constants.BlockSearchPagination
	}
}

func (a *App) PopulateActiveSet() {
	if a.IsPopulatingActiveSet {
		a.Logger.Debug().Msg("App is populating active set already, not populating again")
		return
	}

	a.Logger.Debug().Msg("Populating active set...")

	a.IsPopulatingActiveSet = true

	block, err := a.RPC.GetLatestBlock()
	if err != nil {
		a.Logger.Error().Err(err).Msg("Error querying for last block")
		a.IsPopulatingActiveSet = false
		return
	}

	blockParsed := block.Result.Block.ToBlock()

	a.Logger.Info().
		Int64("height", blockParsed.Height).
		Msg("Last chain block")

	blockHeight := blockParsed.Height

	for {
		count := a.StateManager.GetActiveSetsCountSinceLatest(a.Config.ChainConfig.StoreBlocks)
		if count >= a.Config.ChainConfig.StoreBlocks {
			a.Logger.Info().
				Int64("count", count).
				Msg("Got enough historical validators for populating")
			a.IsPopulatingActiveSet = false
			break
		}

		if a.StateManager.HasActiveSetAtHeight(blockHeight) {
			a.Logger.Trace().
				Int64("height", blockHeight).
				Msg("Already have active set at this block, skipping")
			blockHeight -= 1
			continue
		}

		a.Logger.Info().
			Int64("count", count).
			Int64("required", a.Config.ChainConfig.StoreBlocks).
			Msg("Not enough historical validators, fetching more...")

		heightActiveSet, err := a.RPC.GetActiveSetAtBlock(blockHeight)
		if err != nil {
			a.Logger.Error().
				Err(err).
				Int64("height", blockHeight).
				Msg("Error querying for active set at height")
			a.IsPopulatingActiveSet = false
			return
		}

		if err := a.StateManager.AddActiveSet(blockHeight, heightActiveSet); err != nil {
			a.Logger.Error().
				Err(err).
				Msg("Error inserting active set")
			a.IsPopulatingActiveSet = false
			return
		}

		blockHeight -= 1
	}
}
