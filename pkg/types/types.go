package types

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

type Block struct {
	Height     int64
	Time       time.Time
	Signatures []Signature
}

func (b *Block) Hash() string {
	return fmt.Sprintf("block_%d", b.Height)
}

type Signature struct {
	Height        int64
	ValidatorAddr string
	Signed        bool
}

type WebsocketEmittable interface {
	Hash() string
}

type WSError struct {
	Error error
}

func (w *WSError) Hash() string {
	return "error_" + uuid.NewString()
}
