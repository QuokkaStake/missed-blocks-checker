package types

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWsErrorHash(t *testing.T) {
	t.Parallel()

	err := WSError{Error: errors.New("error")}
	assert.True(t, strings.HasPrefix(err.Hash(), "error_"), "Wrong error hash!")
}
