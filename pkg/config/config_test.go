package config_test

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	"main/pkg/fs"
	"testing"

	"github.com/stretchr/testify/require"
)

type TmpFSInterface struct {
}

func (filesystem *TmpFSInterface) ReadFile(name string) ([]byte, error) {
	return assets.EmbedFS.ReadFile(name)
}

func (filesystem *TmpFSInterface) Create(path string) (fs.File, error) {
	return nil, errors.New("not yet supported")
}

func TestLoadConfigErrorReading(t *testing.T) {
	t.Parallel()

	config, err := configPkg.GetConfig("nonexistent.toml", &TmpFSInterface{})

	require.Error(t, err)
	require.Nil(t, config)
}

func TestLoadConfigInvalidToml(t *testing.T) {
	t.Parallel()

	config, err := configPkg.GetConfig("invalid.toml", &TmpFSInterface{})

	require.Error(t, err)
	require.Nil(t, config)
}

func TestLoadConfigValid(t *testing.T) {
	t.Parallel()

	config, err := configPkg.GetConfig("valid.toml", &TmpFSInterface{})

	require.NoError(t, err)
	require.NotNil(t, config)
}

func TestValidateConfigEmptyChains(t *testing.T) {
	t.Parallel()

	config := configPkg.Config{ChainConfigs: []*configPkg.ChainConfig{}}
	err := config.Validate()

	require.Error(t, err)
}

func TestValidateConfigInvalidChain(t *testing.T) {
	t.Parallel()

	config := configPkg.Config{ChainConfigs: []*configPkg.ChainConfig{
		{},
	}}
	err := config.Validate()

	require.Error(t, err)
}

func TestValidateConfigInvalidDatabase(t *testing.T) {
	t.Parallel()

	config := configPkg.Config{
		ChainConfigs: []*configPkg.ChainConfig{
			{
				Name:         "chain",
				FetcherType:  "cosmos-rpc",
				RPCEndpoints: []string{"https://example.com"},
				Thresholds:   []float64{0, 50, 100},
				EmojisStart:  []string{"x", "y"},
				EmojisEnd:    []string{"a", "b"},
			},
		},
		DatabaseConfig: configPkg.DatabaseConfig{Type: "wrong"},
	}
	err := config.Validate()

	require.Error(t, err)
}

func TestValidateConfigValid(t *testing.T) {
	t.Parallel()

	config := configPkg.Config{
		ChainConfigs: []*configPkg.ChainConfig{
			{
				Name:         "chain",
				FetcherType:  "cosmos-rpc",
				RPCEndpoints: []string{"https://example.com"},
				Thresholds:   []float64{0, 50, 100},
				EmojisStart:  []string{"x", "y"},
				EmojisEnd:    []string{"a", "b"},
			},
		},
		DatabaseConfig: configPkg.DatabaseConfig{Type: "sqlite", Path: "sqlite.sql"},
	}
	err := config.Validate()

	require.NoError(t, err)
}
