package converter

import (
	"main/pkg/types"
	"testing"
	"time"

	"cosmossdk.io/math"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
)

func TestConverterFromCosmosValidator(t *testing.T) {
	t.Parallel()

	converter := NewConverter()
	src := stakingTypes.Validator{
		OperatorAddress: "cosmosvaloper1qphf0ferqcch0jca9hlqfm3x0eds3dpkcvpafp",
		ConsensusPubkey: &codecTypes.Any{
			TypeUrl: "/cosmos.crypto.ed25519.PubKey",
			Value:   []byte{10, 32, 190, 133, 104, 92, 29, 0, 175, 54, 121, 236, 216, 25, 131, 32, 33, 175, 6, 180, 153, 86, 155, 62, 40, 222, 169, 61, 30, 109, 2, 88, 60, 247},
		},
		Jailed:          false,
		Status:          stakingTypes.Unbonded,
		Tokens:          math.NewInt(1386400),
		DelegatorShares: math.LegacyNewDec(1386400.000000000000000000),
		Description: stakingTypes.Description{
			Moniker:         "moniker",
			Identity:        "identity",
			Website:         "website",
			SecurityContact: "contact",
			Details:         "details",
		},
		UnbondingHeight: 0,
		UnbondingTime:   time.Now(),
		Commission: stakingTypes.Commission{
			CommissionRates: stakingTypes.CommissionRates{
				Rate:          math.LegacyMustNewDecFromStr("0.1"),
				MaxRate:       math.LegacyMustNewDecFromStr("0.2"),
				MaxChangeRate: math.LegacyMustNewDecFromStr("0.01"),
			},
			UpdateTime: time.Now(),
		},
		UnbondingOnHoldRefCount: 0,
	}

	val := converter.ValidatorFromCosmosValidator(src, &slashingTypes.ValidatorSigningInfo{
		Tombstoned:          false,
		MissedBlocksCounter: 10,
	})

	assert.Equal(t, &types.Validator{
		Moniker:                 "moniker",
		Description:             "details",
		Identity:                "identity",
		SecurityContact:         "contact",
		Website:                 "website",
		ConsensusAddressHex:     "E5464CB88318A98724BFFE9E0C59129A2B35F11E",
		ConsensusAddressValcons: "cosmosvalcons1u4ryewyrrz5cwf9ll60qckgjng4ntug726d6vf",
		OperatorAddress:         "cosmosvaloper1qphf0ferqcch0jca9hlqfm3x0eds3dpkcvpafp",
		Commission:              0.1,
		Jailed:                  false,
		SigningInfo: &types.SigningInfo{
			MissedBlocksCounter: 10,
			Tombstoned:          false,
		},
		VotingPower:                  math.LegacyNewDec(1386400.000000000000000000),
		VotingPowerPercent:           0,
		CumulativeVotingPowerPercent: 0,
		Rank:                         0,
	}, val)

	converter.MustSetValidatorConsumerConsensusAddr(val, "cosmosvalcons16vj0vma5lzcfsdfnkl7rrk9tfke9ut5yesmumm")
	assert.Equal(t, "D324F66FB4F8B0983533B7FC31D8AB4DB25E2E84", val.ConsensusAddressHex)
	assert.Equal(t, "cosmosvalcons16vj0vma5lzcfsdfnkl7rrk9tfke9ut5yesmumm", val.ConsensusAddressValcons)
}
