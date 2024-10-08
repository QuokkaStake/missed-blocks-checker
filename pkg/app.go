package pkg

import (
	configPkg "main/pkg/config"
	databasePkg "main/pkg/database"
	"main/pkg/fs"
	loggerPkg "main/pkg/logger"
	"main/pkg/metrics"

	"github.com/rs/zerolog"
)

type App struct {
	Logger         zerolog.Logger
	Config         *configPkg.Config
	Database       *databasePkg.Database
	MetricsManager *metrics.Manager
	Version        string

	AppManagers []*AppManager
}

func NewApp(configPath string, filesystem fs.FS, version string) *App {
	config, err := configPkg.GetConfig(configPath, filesystem)
	if err != nil {
		loggerPkg.GetDefaultLogger().Panic().Err(err).Msg("Could not load config")
	}

	if err = config.Validate(); err != nil {
		loggerPkg.GetDefaultLogger().Panic().Err(err).Msg("Provided config is invalid!")
	}

	for _, chainConfig := range config.ChainConfigs {
		chainConfig.RecalculateMissedBlocksGroups()
	}

	logger := loggerPkg.GetLogger(config.LogConfig).
		With().
		Str("component", "app_manager").
		Logger()

	metricsManager := metrics.NewManager(logger, config.MetricsConfig)
	database := databasePkg.NewDatabase(logger, config.DatabaseConfig)

	appManagers := make([]*AppManager, len(config.ChainConfigs))
	for index, chainConfig := range config.ChainConfigs {
		appManagers[index] = NewAppManager(
			logger,
			chainConfig,
			version,
			metricsManager,
			database,
		)
	}

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

	for _, chainConfig := range a.Config.ChainConfigs {
		a.MetricsManager.SetDefaultMetrics(chainConfig)
	}
	a.MetricsManager.LogAppVersion(a.Version)

	for _, appManager := range a.AppManagers {
		go appManager.Start()
	}

	select {}
}
