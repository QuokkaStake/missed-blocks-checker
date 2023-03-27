package types

import (
	"github.com/google/uuid"
)

type WSError struct {
	Error error
}

func (w *WSError) Hash() string {
	return "error_" + uuid.NewString()
}
