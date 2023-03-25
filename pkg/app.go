package pkg

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/logger"
	statePkg "main/pkg/state"
	"main/pkg/tendermint"
	"main/pkg/types"

	"github.com/rs/zerolog"
)

type App struct {
	Logger       *zerolog.Logger
	Config       *configPkg.Config
	RPC          *tendermint.RPC
	StateManager *statePkg.Manager
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

	return &App{
		Logger:       log,
		Config:       config,
		RPC:          rpc,
		StateManager: stateManager,
	}
}

func (a *App) Start() {
	a.StateManager.Init()
	a.UpdateValidators()

	go a.ListenForEvents()
	a.Populate()

	select {}
}

func (a *App) ListenForEvents() {
	wsClient := tendermint.NewWebsocketClient(a.Logger, a.Config.ChainConfig.RPCEndpoints[0], a.Config)
	go wsClient.Listen()

	for {
		select {
		case result := <-wsClient.Channel:
			block, ok := result.(*types.Block)
			if !ok {
				a.Logger.Warn().Msg("Event is not a block!")
				continue
			}

			a.StateManager.AddBlock(block)
			count := a.StateManager.GetBlocksCountSinceLatest(constants.StoreBlocks)

			if count < constants.StoreBlocks {
				continue
			}

			snapshot := a.StateManager.GetSnapshot()

			for _, entry := range snapshot.Entries {
				a.Logger.Info().
					Str("valoper", entry.OperatorAddress).
					Str("moniker", entry.Moniker).
					Int64("signed", entry.SignatureInfo.Signed).
					Int64("not_signed", entry.SignatureInfo.NotSigned).
					Int64("no_signature", entry.SignatureInfo.NoSignature).
					Int64("proposed", entry.SignatureInfo.Proposed).
					Msg("Validator signing info")
			}

			a.Logger.Info().
				Int64("count", count).
				Int64("height", block.Height).
				Msg("Added blocks into state")
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

func (a *App) Populate() {
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
		count := a.StateManager.GetBlocksCountSinceLatest(constants.StoreBlocks)
		if count >= constants.StoreBlocks {
			a.Logger.Info().
				Int64("count", count).
				Msg("Got enough blocks for populating")
			break
		}

		a.Logger.Info().
			Int64("count", count).
			Int64("required", constants.StoreBlocks).
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
