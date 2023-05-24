package config

import "fmt"

type DatabaseConfig struct {
	Path string `toml:"path"`
}

func (c *DatabaseConfig) Validate() error {
	if c.Path == "" {
		return fmt.Errorf("database path not specified")
	}

	return nil
}
