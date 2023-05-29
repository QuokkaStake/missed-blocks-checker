package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMissedBlocksGroupEmpty(t *testing.T) {
	t.Parallel()

	groups := MissedBlocksGroups{}
	err := groups.Validate(10000)
	assert.NotNil(t, err, "Error should be present!")
}

func TestMissedBlocksGroupMissingStart(t *testing.T) {
	t.Parallel()

	groups := MissedBlocksGroups{
		{Start: 5000, End: 10000},
	}
	err := groups.Validate(10000)
	assert.NotNil(t, err, "Error should be present!")
}

func TestMissedBlocksGroupMissingEnd(t *testing.T) {
	t.Parallel()

	groups := MissedBlocksGroups{
		{Start: 0, End: 5000},
	}
	err := groups.Validate(10000)
	assert.NotNil(t, err, "Error should be present!")
}

func TestMissedBlocksGroupGaps(t *testing.T) {
	t.Parallel()

	groups := MissedBlocksGroups{
		{Start: 0, End: 1000},
		{Start: 9000, End: 10000},
	}
	err := groups.Validate(10000)
	assert.NotNil(t, err, "Error should be present!")
}

func TestMissedBlocksValid(t *testing.T) {
	t.Parallel()

	groups := MissedBlocksGroups{
		{Start: 0, End: 4999},
		{Start: 5000, End: 10000},
	}
	err := groups.Validate(10000)
	assert.Nil(t, err, "Error should not be present!")
}
