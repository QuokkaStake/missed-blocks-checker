package types

import (
	"fmt"
	"time"
)

type Block struct {
	Height     int64
	Time       time.Time
	Proposer   string
	Signatures map[string]int32
	Validators map[string]bool
}

func (b *Block) Hash() string {
	return fmt.Sprintf("block_%d", b.Height)
}

func (b *Block) SetValidators(validators map[string]bool) {
	b.Validators = validators
}
