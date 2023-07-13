package config

type LogConfig struct {
	LogLevel   string `default:"info"  toml:"level"`
	JSONOutput bool   `default:"false" toml:"json"`
}
