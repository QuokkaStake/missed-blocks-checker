package types

type WebsocketEmittable interface {
	Hash() string
}

type Link struct {
	Href string
	Text string
}

type BlocksMap map[int64]*Block
type Validators []*Validator
type ValidatorsMap map[string]*Validator

func (validators Validators) ToMap() ValidatorsMap {
	validatorsMap := make(ValidatorsMap, len(validators))

	for _, validator := range validators {
		validatorsMap[validator.OperatorAddress] = validator
	}

	return validatorsMap
}

func (validatorsMap ValidatorsMap) ToSlice() Validators {
	validators := make(Validators, len(validatorsMap))
	index := 0

	for _, validator := range validatorsMap {
		validators[index] = validator
		index++
	}

	return validators
}
