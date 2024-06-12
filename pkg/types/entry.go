package types

type Entry struct {
	IsActive      bool
	Validator     *Validator
	SignatureInfo SignatureInto
}
