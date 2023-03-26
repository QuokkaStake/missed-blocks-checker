package pkg

import (
	"fmt"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/logger"
	statePkg "main/pkg/state"
	"main/pkg/tendermint"
	"main/pkg/types"
	"time"

	"github.com/rs/zerolog"
)

type App struct {
	Logger           *zerolog.Logger
	Config           *configPkg.Config
	RPC              *tendermint.RPC
	StateManager     *statePkg.Manager
	WebsocketManager *tendermint.WebsocketManager
	IsPopulating     bool
}

func NewApp(configPath string) *App {
	config, err := configPkg.GetConfig(configPath)
	if err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not load config")
	}

	if err = config.Validate(); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Provided config is invalid!")
	}

	log := logger.GetLogger(config.LogConfig)
	rpc := tendermint.NewRPC(config.ChainConfig.RPCEndpoints, log)
	stateManager := statePkg.NewManager(log, config)
	websocketManager := tendermint.NewWebsocketManager(log, config)

	return &App{
		Logger:           log,
		Config:           config,
		RPC:              rpc,
		StateManager:     stateManager,
		WebsocketManager: websocketManager,
		IsPopulating:     false,
	}
}

func (a *App) Start() {
	a.StateManager.Init()
	a.UpdateValidators()

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

			a.Logger.Debug().Int64("height", block.Height).Msg("Got new block from Tendermint")
			a.StateManager.AddBlock(block)

			count := a.StateManager.GetBlocksCountSinceLatest(a.Config.ChainConfig.StoreBlocks)

			a.Logger.Info().
				Int64("count", count).
				Int64("height", block.Height).
				Msg("Added blocks into state")

			if count < a.Config.ChainConfig.StoreBlocks {
				a.Logger.Debug().
					Int64("count", count).
					Int64("expected", a.Config.ChainConfig.BlocksWindow).
					Msg("Not enough blocks for producing a snapshot, skipping.")
				continue
			}

			snapshot := a.StateManager.GetSnapshot()

			for _, entry := range snapshot.Entries {
				a.Logger.Debug().
					Str("valoper", entry.OperatorAddress).
					Str("moniker", entry.Moniker).
					Int64("signed", entry.SignatureInfo.Signed).
					Int64("not_signed", entry.SignatureInfo.NotSigned).
					Int64("no_signature", entry.SignatureInfo.NoSignature).
					Int64("proposed", entry.SignatureInfo.Proposed).
					Msg("Validator signing info")
			}

			if olderSnapshot == nil {
				a.Logger.Info().Msg("No older snapshot present, cannot generate snapshot diff")
				olderSnapshot = snapshot
				continue
			}

			diff := snapshot.GetDiff(olderSnapshot)
			olderSnapshot = snapshot

			if diff.Empty() {
				a.Logger.Debug().Msg("Snapshot diff is empty, no events to send.")
				continue
			}

			for _, entry := range diff.Entries {
				a.Logger.Debug().
					Str("diff", fmt.Sprintf("%+v", entry)).
					Msg("Snapshot diff entry")
			}
		}
	}
}

func (a *App) UpdateValidators() error {
	validators, err := a.RPC.GetValidators()
	if err != nil {
		return err

	}

	a.StateManager.State.SetValidators(validators)
	return nil
}

func (a *App) PopulateInBackground() {
	a.Populate()

	ticker := time.NewTicker(60 * time.Second)
	quit := make(chan struct{})

	for {
		select {
		case <-ticker.C:
			a.Populate()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func (a *App) Populate() {
	if a.IsPopulating {
		a.Logger.Debug().Msg("App is populating already, not populating again")
		return
	}

	a.IsPopulating = true

	block, err := a.RPC.GetLatestBlock()
	if err != nil {
		a.Logger.Fatal().Err(err).Msg("Error querying for last block")
	}

	blockParsed := block.Result.Block.ToBlock()

	a.Logger.Info().
		Int64("height", blockParsed.Height).
		Msg("Last chain block")

	a.StateManager.AddBlock(blockParsed)

	startBlockToFetch := blockParsed.Height

	for {
		count := a.StateManager.GetBlocksCountSinceLatest(a.Config.ChainConfig.StoreBlocks)
		if count >= a.Config.ChainConfig.StoreBlocks {
			a.Logger.Info().
				Int64("count", count).
				Msg("Got enough blocks for populating")
			a.IsPopulating = false
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
			a.Logger.Fatal().Err(err).Msg("Error querying for blocks search")
		}

		for _, block := range blocks.Result.Blocks {
			a.StateManager.AddBlock(block.Block.ToBlock())
		}

		startBlockToFetch -= constants.BlockSearchPagination
	}
}
