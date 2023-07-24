package types

type SignatureInto struct {
	BlocksCount int64
	Signed      int64
	NoSignature int64
	NotSigned   int64
	NotActive   int64
	Active      int64
	Proposed    int64
}

func (s *SignatureInto) GetNotSigned() int64 {
	return s.NotSigned + s.NoSignature
}
