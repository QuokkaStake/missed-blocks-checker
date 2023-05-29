package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateDatabaseConfigWithoutPath(t *testing.T) {
	t.Parallel()

	config := &DatabaseConfig{}
	err := config.Validate()
	assert.NotNil(t, err, "Error should be present!")
}

func TestValidateDatabaseConfigOk(t *testing.T) {
	t.Parallel()

	config := &DatabaseConfig{Path: "path"}
	err := config.Validate()
	assert.Nil(t, err, "Error should not be present!")
}
