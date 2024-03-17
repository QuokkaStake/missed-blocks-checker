package config_test

import (
	"main/pkg/config"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExplorerGetValidatorLinkMintscan(t *testing.T) {
	t.Parallel()

	explorer := &config.ExplorerConfig{MintscanPrefix: "test"}
	link := explorer.GetValidatorLink(&types.Validator{OperatorAddress: "address", Moniker: "moniker"})
	require.Equal(t, "moniker", link.Text)
	require.Equal(t, "https://mintscan.io/test/validators/address", link.Href)
}

func TestExplorerGetValidatorLinkPingPub(t *testing.T) {
	t.Parallel()

	explorer := &config.ExplorerConfig{PingPubPrefix: "test", PingPubHost: "https://example.com"}
	link := explorer.GetValidatorLink(&types.Validator{OperatorAddress: "address", Moniker: "moniker"})
	require.Equal(t, "moniker", link.Text)
	require.Equal(t, "https://example.com/test/staking/address", link.Href)
}

func TestExplorerGetValidatorLinkCustom(t *testing.T) {
	t.Parallel()

	explorer := &config.ExplorerConfig{ValidatorLinkPattern: "https://example.com/validator/%s"}
	link := explorer.GetValidatorLink(&types.Validator{OperatorAddress: "address", Moniker: "moniker"})
	require.Equal(t, "moniker", link.Text)
	require.Equal(t, "https://example.com/validator/address", link.Href)
}

func TestExplorerGetValidatorLinkNoExplorer(t *testing.T) {
	t.Parallel()

	explorer := &config.ExplorerConfig{}
	link := explorer.GetValidatorLink(&types.Validator{OperatorAddress: "address", Moniker: "moniker"})
	require.Equal(t, "moniker", link.Text)
	require.Equal(t, "", link.Href)
}
