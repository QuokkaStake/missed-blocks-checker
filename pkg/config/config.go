package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/creasty/defaults"
)

type Config struct {
	LogConfig      LogConfig      `toml:"log"`
	ChainConfigs   []*ChainConfig `toml:"chains"`
	DatabaseConfig DatabaseConfig `toml:"database"`
	MetricsConfig  MetricsConfig  `toml:"metrics"`
}

func (config *Config) Validate() error {
	if len(config.ChainConfigs) == 0 {
		return fmt.Errorf("no chains specified")
	}

	for index, chainConfig := range config.ChainConfigs {
		if err := chainConfig.Validate(); err != nil {
			return fmt.Errorf("error in chain config #%d: %s", index, err)
		}
	}

	if err := config.DatabaseConfig.Validate(); err != nil {
		return fmt.Errorf("error in database config: %s", err)
	}

	return nil
}

func GetConfig(path string) (*Config, error) {
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	configString := string(configBytes)

	configStruct := &Config{}
	if _, err = toml.Decode(configString, configStruct); err != nil {
		return nil, err
	}
	if err := defaults.Set(configStruct); err != nil {
		return nil, err
	}

	return configStruct, nil
}
