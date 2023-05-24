package config

import (
	"fmt"
	"main/pkg/types"
)

type ExplorerConfig struct {
	ValidatorLinkPattern string `toml:"validator-link-pattern"`
	MintscanPrefix       string `toml:"mintscan-prefix"`
}

func (c *ExplorerConfig) GetValidatorLink(validator *types.Validator) types.Link {
	if c.MintscanPrefix != "" {
		return types.Link{
			Href: fmt.Sprintf(
				"https://mintscan.io/%s/validators/%s",
				c.MintscanPrefix,
				validator.OperatorAddress,
			),
			Text: validator.Moniker,
		}
	}

	if c.ValidatorLinkPattern != "" {
		return types.Link{
			Href: fmt.Sprintf(c.ValidatorLinkPattern, validator.OperatorAddress),
			Text: validator.Moniker,
		}
	}

	return types.Link{Text: validator.Moniker}
}
