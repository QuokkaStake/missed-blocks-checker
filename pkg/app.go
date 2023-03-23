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
	Logger *zerolog.Logger
	Config *configPkg.Config
	RPC    *tendermint.RPC
	State  *statePkg.State
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

	return &App{
		Logger: log,
		Config: config,
		RPC:    rpc,
		State:  statePkg.NewState(),
	}
}

func (a *App) Start() {
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

			a.State.AddBlock(block)
			count := a.State.GetBlocksCountSinceLatest(constants.StoreBlocks)

			a.Logger.Info().
				Int64("count", count).
				Msg("Added blocks into state")
		}
	}
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

	a.State.AddBlock(blockParsed)

	startBlockToFetch := blockParsed.Height

	for {
		count := a.State.GetBlocksCountSinceLatest(constants.StoreBlocks)
		if count >= 10000 {
			a.Logger.Info().Msg("Got enough blocks for populating")
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
			a.State.AddBlock(block.Block.ToBlock())
		}

		startBlockToFetch -= constants.BlockSearchPagination
	}
}
