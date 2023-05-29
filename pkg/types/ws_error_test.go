package types

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestWsErrorHash(t *testing.T) {
	t.Parallel()

	err := WSError{Error: fmt.Errorf("error")}
	assert.True(t, strings.HasPrefix(err.Hash(), "error_"), "Wrong error hash!")
}
