package config

import (
	"os"

	"github.com/BurntSushi/toml"
	defaults "github.com/mcuadros/go-defaults"
)

type Config struct {
	TelegramConfig TelegramConfig `toml:"telegram"`
	LogConfig      LogConfig      `toml:"log"`
	ChainConfig    ChainConfig    `toml:"chain"`
}

type ChainConfig struct {
	Name         string   `toml:"name"`
	RPCEndpoints []string `toml:"rpc-endpoints"`
}

type TelegramConfig struct {
	TelegramChat  int64  `toml:"chat"`
	TelegramToken string `toml:"token"`
}

type LogConfig struct {
	LogLevel   string `toml:"level" default:"info"`
	JSONOutput bool   `toml:"json" default:"false"`
}

func (c *Config) Validate() error {
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
	defaults.SetDefaults(configStruct)

	return configStruct, nil
}
