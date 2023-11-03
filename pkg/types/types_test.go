package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorsToMap(t *testing.T) {
	t.Parallel()

	validators := Validators{
		{Moniker: "first", OperatorAddress: "firstaddr"},
		{Moniker: "second", OperatorAddress: "secondaddr"},
	}

	validatorsMap := validators.ToMap()
	assert.Len(t, validatorsMap, 2, "Map should have 2 entries!")
	assert.Equal(t, "first", validatorsMap["firstaddr"].Moniker, "Validator mismatch!")
	assert.Equal(t, "second", validatorsMap["secondaddr"].Moniker, "Validator mismatch!")
}

func TestValidatorsToSlice(t *testing.T) {
	t.Parallel()

	validatorsMap := ValidatorsMap{
		"firstaddr":  {Moniker: "first", OperatorAddress: "firstaddr"},
		"secondaddr": {Moniker: "second", OperatorAddress: "secondaddr"},
	}

	validators := validatorsMap.ToSlice()
	assert.Len(t, validators, 2, "Slice should have 2 entries!")

	monikers := []string{
		validators[0].Moniker,
		validators[1].Moniker,
	}

	assert.Contains(t, monikers, "first", "Validator mismatch!")
	assert.Contains(t, monikers, "second", "Validator mismatch!")
}
