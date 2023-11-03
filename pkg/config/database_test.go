package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateDatabaseConfigWithoutPath(t *testing.T) {
	t.Parallel()

	config := &DatabaseConfig{Type: "sqlite"}
	err := config.Validate()
	require.Error(t, err, "Error should be present!")
}

func TestValidateDatabaseConfigWithoutType(t *testing.T) {
	t.Parallel()

	config := &DatabaseConfig{Path: "path"}
	err := config.Validate()
	require.Error(t, err, "Error should be present!")
}

func TestValidateDatabaseConfigOk(t *testing.T) {
	t.Parallel()

	config := &DatabaseConfig{Path: "path", Type: "sqlite"}
	err := config.Validate()
	require.NoError(t, err, "Error should not be present!")
}
