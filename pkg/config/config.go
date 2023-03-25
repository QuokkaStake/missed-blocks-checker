package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	defaults "github.com/mcuadros/go-defaults"
)

type Config struct {
	TelegramConfig TelegramConfig `toml:"telegram"`
	LogConfig      LogConfig      `toml:"log"`
	ChainConfig    ChainConfig    `toml:"chain"`
	DatabaseConfig DatabaseConfig `toml:"database"`
}

type DatabaseConfig struct {
	Path string `toml:"path"`
}

func (c *DatabaseConfig) Validate() error {
	if c.Path == "" {
		return fmt.Errorf("database path not specified")
	}

	return nil
}

type ChainConfig struct {
	Name         string   `toml:"name"`
	RPCEndpoints []string `toml:"rpc-endpoints"`
}

func (c *ChainConfig) Validate() error {
	if len(c.RPCEndpoints) == 0 {
		return fmt.Errorf("chain has 0 RPC endpoints")
	}

	return nil
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
	if err := c.ChainConfig.Validate(); err != nil {
		return fmt.Errorf("error in chain config: %s", err)
	}

	if err := c.DatabaseConfig.Validate(); err != nil {
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
	defaults.SetDefaults(configStruct)

	return configStruct, nil
}
