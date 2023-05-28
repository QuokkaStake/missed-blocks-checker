package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNameWithoutPrettyName(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{Name: "test"}
	assert.Equal(t, config.GetName(), "test", "Names do not match!")
}

func TestGetNameWithPrettyName(t *testing.T) {
	t.Parallel()

	config := &ChainConfig{Name: "test", PrettyName: "Test"}
	assert.Equal(t, config.GetName(), "Test", "Names do not match!")
}
