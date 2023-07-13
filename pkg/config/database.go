package config

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/utils"
	"strings"
)

type DatabaseConfig struct {
	Type string `json:"type"`
	Path string `toml:"path"`
}

func (c *DatabaseConfig) Validate() error {
	types := constants.GetDatabaseTypes()

	if !utils.Contains(types, c.Type) {
		return fmt.Errorf(
			"wrong database type: expected one of %s, but got \"%s\"",
			strings.Join(types, ", "),
			c.Type,
		)
	}

	if c.Path == "" {
		return fmt.Errorf("database path not specified")
	}

	return nil
}
