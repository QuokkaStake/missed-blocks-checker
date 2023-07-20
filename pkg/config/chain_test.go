package config

import (
	"gopkg.in/guregu/null.v4"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetChainNameWithoutPrettyName(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{Name: "test"}
	assert.Equal(t, config.GetName(), "test", "Names do not match!")
}

func TestGetChainNameWithPrettyName(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{Name: "test", PrettyName: "Test"}
	assert.Equal(t, config.GetName(), "Test", "Names do not match!")
}

func TestGetChainBlocksSignCount(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{BlocksWindow: 10000, MinSignedPerWindow: 0.05}
	assert.Equal(t, config.GetBlocksSignCount(), int64(9500), "Blocks sign count does not match!")
}

func TestValidateChainWithoutName(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{}
	err := config.Validate()
	assert.NotNil(t, err, "Error should be present!")
}

func TestValidateChainWithoutRPCEndpoints(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{Name: "chain"}
	err := config.Validate()
	assert.NotNil(t, err, "Error should be present!")
}

func TestValidateConsumerChainWithoutProviderEndpoints(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"endpoint"},
		IsConsumer:   null.BoolFrom(true),
	}
	err := config.Validate()
	assert.NotNil(t, err, "Error should be present!")
}

func TestValidateChainValid(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{Name: "chain", RPCEndpoints: []string{"endpoint"}}
	err := config.Validate()
	assert.Nil(t, err, "Error should not be present!")
}

func TestValidateConsumerChainValid(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{
		Name:                 "chain",
		RPCEndpoints:         []string{"endpoint"},
		IsConsumer:           null.BoolFrom(true),
		ProviderRPCEndpoints: []string{"endpoint"},
	}
	err := config.Validate()
	assert.Nil(t, err, "Error should not be present!")
}
