package pkg

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/logger"
	statePkg "main/pkg/state"
	"main/pkg/tendermint"

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
	a.Populate()

	select {}
}

func (a *App) Populate() {
	block, err := a.RPC.GetLatestBlock()
	if err != nil {
		panic(err)
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
			panic(err)
		}

		for _, block := range blocks.Result.Blocks {
			a.State.AddBlock(block.Block.ToBlock())
		}

		startBlockToFetch -= 100
	}
}
