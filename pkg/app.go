package pkg

import (
	"github.com/rs/zerolog"
	configPkg "main/pkg/config"
	databasePkg "main/pkg/database"
	loggerPkg "main/pkg/logger"
	"main/pkg/metrics"
)

type App struct {
	Logger         zerolog.Logger
	Config         *configPkg.Config
	Database       *databasePkg.Database
	MetricsManager *metrics.Manager
	Version        string

	AppManagers []*AppManager
}

func NewApp(configPath string, version string) *App {
	config, err := configPkg.GetConfig(configPath)
	if err != nil {
		loggerPkg.GetDefaultLogger().Fatal().Err(err).Msg("Could not load config")
	}
	config.ChainConfig.SetDefaultMissedBlocksGroups()

	if err = config.Validate(); err != nil {
		loggerPkg.GetDefaultLogger().Fatal().Err(err).Msg("Provided config is invalid!")
	}

	logger := loggerPkg.GetLogger(config.LogConfig).
		With().
		Str("component", "app_manager").
		Logger()

	metricsManager := metrics.NewManager(logger, config.MetricsConfig)
	database := databasePkg.NewDatabase(logger, config.DatabaseConfig)

	appManagers := make([]*AppManager, 1)
	appManagers[0] = NewAppManager(
		logger,
		config.ChainConfig,
		metricsManager,
		database,
	)

	return &App{
		Logger:         logger,
		Config:         config,
		Database:       database,
		MetricsManager: metricsManager,
		Version:        version,
		AppManagers:    appManagers,
	}
}

func (a *App) Start() {
	a.Database.Init()
	go a.MetricsManager.Start()

	a.MetricsManager.SetDefaultMetrics(a.Config.ChainConfig)
	a.MetricsManager.LogAppVersion(a.Version)

	for _, appManager := range a.AppManagers {
		go appManager.Start()
	}

	select {}
}
