package types

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWsErrorHash(t *testing.T) {
	t.Parallel()

	err := WSError{Error: fmt.Errorf("error")}
	assert.True(t, strings.HasPrefix(err.Hash(), "error_"), "Wrong error hash!")
}
