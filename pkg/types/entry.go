package types

import (
	"main/pkg/utils"
)

type Entry struct {
	IsActive      bool
	Validator     *Validator
	SignatureInfo SignatureInto
}

type Entries map[string]Entry

func (e Entries) ToSlice() []Entry {
	entries := make([]Entry, len(e))

	index := 0
	for _, entry := range e {
		entries[index] = entry
		index++
	}

	return entries
}

func (e Entries) ByValidatorAddresses(addresses []string) []Entry {
	entries := make([]Entry, 0)

	for _, entry := range e {
		if utils.Contains(addresses, entry.Validator.OperatorAddress) {
			entries = append(entries, entry)
		}
	}

	return entries
}
