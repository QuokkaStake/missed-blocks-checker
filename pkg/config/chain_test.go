package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNameWithoutPrettyName(t *testing.T) {
	config := &ChainConfig{Name: "test"}
	assert.Equal(t, config.GetName(), "test", "Names do not match!")
}

func TestGetNameWithPrettyName(t *testing.T) {
	config := &ChainConfig{Name: "test", PrettyName: "Test"}
	assert.Equal(t, config.GetName(), "Test", "Names do not match!")
}
