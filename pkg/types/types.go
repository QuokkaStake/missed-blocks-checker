package types

import "time"

type Block struct {
	Height     int64
	Time       time.Time
	Signatures []Signature
}

type Signature struct {
	Height        int64
	ValidatorAddr string
	Signed        bool
}
