package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/google/uuid"
	"main/pkg/constants"
	"main/pkg/utils"
	"time"
)

type Block struct {
	Height     int64
	Time       time.Time
	Proposer   string
	Signatures map[string]int32
}

func (b *Block) Hash() string {
	return fmt.Sprintf("block_%d", b.Height)
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

type Validator struct {
	Moniker          string
	ConsensusAddress string
	OperatorAddress  string
	Status           int32
	Jailed           bool
}

func (v *Validator) Active() bool {
	return v.Status == constants.ValidatorBonded
}

func ValidatorFromCosmosValidator(validator stakingTypes.Validator) *Validator {
	interfaceRegistry := types.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	parseCodec := codec.NewProtoCodec(interfaceRegistry)

	if err := validator.UnpackInterfaces(parseCodec); err != nil {
		panic(err)
	}

	addr, err := validator.GetConsAddr()
	if err != nil {
		panic(err)
	}

	return &Validator{
		Moniker:          validator.Description.Moniker,
		ConsensusAddress: fmt.Sprintf("%x", addr),
		OperatorAddress:  validator.OperatorAddress,
		Status:           int32(validator.Status),
		Jailed:           validator.Jailed,
	}
}

type SignatureInto struct {
	Signed      int64
	NoSignature int64
	NotSigned   int64
	Proposed    int64
}

func (s *SignatureInto) GetNotSigned() int64 {
	return s.NotSigned + s.NoSignature
}

type Link struct {
	Href string
	Text string
}

// Notifiers has the following struct:
// notifiers[valoper][reporter] == [reporter1, reporter2]
type Notifiers map[string]map[string][]string

func (n Notifiers) AddNotifier(operatorAddress, reporter, notifier string) bool {
	if _, ok := n[operatorAddress]; !ok {
		n[operatorAddress] = make(map[string][]string)
	}

	if _, ok := n[operatorAddress][reporter]; !ok {
		n[operatorAddress][reporter] = make([]string, 1)
	}

	if utils.Contains(n[operatorAddress][reporter], notifier) {
		return false
	}

	n[operatorAddress][reporter] = append(n[operatorAddress][reporter], notifier)
	return true
}

func (n Notifiers) GetNotifiers(operatorAddress, reporter string) []string {
	if _, ok := n[operatorAddress]; !ok {
		return []string{}
	}

	if _, ok := n[operatorAddress][reporter]; !ok {
		return []string{}
	}

	return n[operatorAddress][reporter]
}
