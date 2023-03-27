package config

import (
	"fmt"
	"main/pkg/types"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/mcuadros/go-defaults"
)

type Config struct {
	TelegramConfig TelegramConfig `toml:"telegram"`
	LogConfig      LogConfig      `toml:"log"`
	ChainConfig    ChainConfig    `toml:"chain"`
	DatabaseConfig DatabaseConfig `toml:"database"`
	ExplorerConfig ExplorerConfig `toml:"explorer"`
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

type ExplorerConfig struct {
	ValidatorLinkPattern string `toml:"validator-link-pattern"`
	MintscanPrefix       string `toml:"mintscan-prefix"`
}

func (c *ExplorerConfig) GetValidatorLink(validator *types.Validator) types.Link {
	if c.MintscanPrefix != "" {
		return types.Link{
			Href: fmt.Sprintf(
				"https://mintscan.io/%s/validators/%s",
				c.MintscanPrefix,
				validator.OperatorAddress,
			),
			Text: validator.Moniker,
		}
	}

	if c.ValidatorLinkPattern != "" {
		return types.Link{
			Href: fmt.Sprintf(c.ValidatorLinkPattern, validator.OperatorAddress),
			Text: validator.Moniker,
		}
	}

	return types.Link{Text: validator.Moniker}
}

type ChainConfig struct {
	Name         string   `toml:"name"`
	RPCEndpoints []string `toml:"rpc-endpoints"`
	StoreBlocks  int64    `toml:"store-blocks" default:"20000"`
	BlocksWindow int64    `toml:"blocks-window" default:"10000"`
}

func (c *ChainConfig) Validate() error {
	if len(c.RPCEndpoints) == 0 {
		return fmt.Errorf("chain has 0 RPC endpoints")
	}

	return nil
}

type TelegramConfig struct {
	Chat   int64   `toml:"chat"`
	Token  string  `toml:"token"`
	Admins []int64 `toml:"admins"`
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
