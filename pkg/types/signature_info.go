package types

type SignatureInto struct {
	Signed      int64
	NoSignature int64
	NotSigned   int64
	Proposed    int64
}

func (s *SignatureInto) GetNotSigned() int64 {
	return s.NotSigned + s.NoSignature
}
