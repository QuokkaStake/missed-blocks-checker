package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gopkg.in/guregu/null.v4"

	"github.com/stretchr/testify/assert"
)

func TestGetChainNameWithoutPrettyName(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{Name: "test"}
	assert.Equal(t, "test", config.GetName(), "Names do not match!")
}

func TestGetChainNameWithPrettyName(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{Name: "test", PrettyName: "Test"}
	assert.Equal(t, "Test", config.GetName(), "Names do not match!")
}

func TestGetChainBlocksSignCount(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{BlocksWindow: 10000, MinSignedPerWindow: 0.05}
	assert.Equal(t, int64(9500), config.GetBlocksSignCount(), "Blocks sign count does not match!")
}

func TestGetChainBlocksMissCount(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{BlocksWindow: 10000, MinSignedPerWindow: 0.05}
	assert.Equal(t, int64(500), config.GetBlocksMissCount(), "Blocks miss count does not match!")
}

func TestValidateChainWithoutName(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{}
	err := config.Validate()
	require.Error(t, err, "Error should be present!")
}

func TestValidateChainWithoutRPCEndpoints(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{Name: "chain"}
	err := config.Validate()
	require.Error(t, err, "Error should be present!")
}

func TestValidateConsumerChainWithoutProviderEndpoints(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"endpoint"},
		IsConsumer:   null.BoolFrom(true),
	}
	err := config.Validate()
	require.Error(t, err, "Error should be present!")
}

func TestValidateConsumerChainWithoutChainID(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{
		Name:                 "chain",
		RPCEndpoints:         []string{"endpoint"},
		IsConsumer:           null.BoolFrom(true),
		ProviderRPCEndpoints: []string{"endpoint"},
	}
	err := config.Validate()
	require.Error(t, err, "Error should be present!")
}

func TestValidateNotEnoughThresholds(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{
		Name: "chain", RPCEndpoints: []string{"endpoint"},
		Thresholds:  []float64{0, 100},
		EmojisStart: []string{"x"},
		EmojisEnd:   []string{"x"},
	}
	err := config.Validate()
	require.Error(t, err, "Error should be present!")
}

func TestValidateNotEnoughStartEmojis(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{
		Name: "chain", RPCEndpoints: []string{"endpoint"},
		Thresholds:  []float64{0, 50, 100},
		EmojisStart: []string{"x"},
		EmojisEnd:   []string{"x", "y"},
	}
	err := config.Validate()
	require.Error(t, err, "Error should be present!")
}

func TestValidateNotEnoughEndEmojis(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{
		Name: "chain", RPCEndpoints: []string{"endpoint"},
		Thresholds:  []float64{0, 50, 100},
		EmojisStart: []string{"x", "y"},
		EmojisEnd:   []string{"x"},
	}
	err := config.Validate()
	require.Error(t, err, "Error should be present!")
}

func TestValidateFirstThresholdNotZero(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{
		Name: "chain", RPCEndpoints: []string{"endpoint"},
		Thresholds:  []float64{1, 50, 100},
		EmojisStart: []string{"x", "y"},
		EmojisEnd:   []string{"x", "y"},
	}
	err := config.Validate()
	require.Error(t, err, "Error should be present!")
}

func TestValidateLastThresholdNotHundred(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{
		Name: "chain", RPCEndpoints: []string{"endpoint"},
		Thresholds:  []float64{0, 50, 95},
		EmojisStart: []string{"x", "y"},
		EmojisEnd:   []string{"x", "y"},
	}
	err := config.Validate()
	require.Error(t, err, "Error should be present!")
}

func TestValidateThresholdsInconsistent(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{
		Name: "chain", RPCEndpoints: []string{"endpoint"},
		Thresholds:  []float64{0, 75, 25, 100},
		EmojisStart: []string{"x", "y", "z"},
		EmojisEnd:   []string{"x", "y", "z"},
	}
	err := config.Validate()
	require.Error(t, err, "Error should be present!")
}

func TestValidateChainValid(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{
		Name: "chain", RPCEndpoints: []string{"endpoint"},
		Thresholds:  []float64{0, 50, 100},
		EmojisStart: []string{"x", "y"},
		EmojisEnd:   []string{"x", "y"},
	}
	err := config.Validate()
	require.NoError(t, err, "Error should not be present!")
}

func TestValidateConsumerChainWithoutEndpoints(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{
		Name:                 "chain",
		RPCEndpoints:         []string{"endpoint"},
		IsConsumer:           null.BoolFrom(true),
		ProviderRPCEndpoints: []string{},
		ConsumerChainID:      "chain",
		Thresholds:           []float64{0, 50, 100},
		EmojisStart:          []string{"x", "y"},
		EmojisEnd:            []string{"x", "y"},
	}
	err := config.Validate()
	require.Error(t, err, "Error should be present!")
}

func TestValidateConsumerChainWithoutChainId(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{
		Name:                 "chain",
		RPCEndpoints:         []string{"endpoint"},
		IsConsumer:           null.BoolFrom(true),
		ProviderRPCEndpoints: []string{"endpoint"},
		ConsumerChainID:      "",
		Thresholds:           []float64{0, 50, 100},
		EmojisStart:          []string{"x", "y"},
		EmojisEnd:            []string{"x", "y"},
	}
	err := config.Validate()
	require.Error(t, err, "Error should be present!")
}

func TestValidateConsumerChainValid(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{
		Name:                 "chain",
		RPCEndpoints:         []string{"endpoint"},
		IsConsumer:           null.BoolFrom(true),
		ProviderRPCEndpoints: []string{"endpoint"},
		ConsumerChainID:      "chain",
		Thresholds:           []float64{0, 50, 100},
		EmojisStart:          []string{"x", "y"},
		EmojisEnd:            []string{"x", "y"},
	}
	err := config.Validate()
	require.NoError(t, err, "Error should not be present!")
}
