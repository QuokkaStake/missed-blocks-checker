package snapshot

import (
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/events"
	"main/pkg/types"
	"main/pkg/utils"
	"math"
	"sort"

	"golang.org/x/exp/slices"
)

type Snapshot struct {
	Entries types.Entries
}

func (snapshot *Snapshot) GetReport(
	olderSnapshot Snapshot,
	chainConfig *config.ChainConfig,
) (*types.Report, error) {
	var entries []types.ReportEvent

	for valoper, entry := range snapshot.Entries {
		olderEntry, ok := olderSnapshot.Entries[valoper]
		if !ok {
			entries = append(entries, events.ValidatorCreated{
				Validator: entry.Validator,
			})
			continue
		}

		hasOlderSigningInfo := olderEntry.Validator.SigningInfo != nil
		hasNewerSigningInfo := entry.Validator.SigningInfo != nil

		if hasOlderSigningInfo &&
			hasNewerSigningInfo &&
			!olderEntry.Validator.SigningInfo.Tombstoned &&
			entry.Validator.SigningInfo.Tombstoned {
			entries = append(entries, events.ValidatorTombstoned{
				Validator: entry.Validator,
			})
			continue
		}

		if entry.Validator.Jailed && !olderEntry.Validator.Jailed && olderEntry.IsActive {
			entries = append(entries, events.ValidatorJailed{
				Validator: entry.Validator,
			})
		}

		if !entry.Validator.Jailed && olderEntry.Validator.Jailed {
			entries = append(entries, events.ValidatorUnjailed{
				Validator: entry.Validator,
			})
		}

		if entry.IsActive && olderEntry.IsActive && entry.Validator.NeedsToSign && !olderEntry.Validator.NeedsToSign {
			entries = append(entries, events.ValidatorJoinedSignatory{
				Validator: entry.Validator,
			})
		}

		if entry.IsActive && olderEntry.IsActive && !entry.Validator.NeedsToSign && olderEntry.Validator.NeedsToSign {
			entries = append(entries, events.ValidatorLeftSignatory{
				Validator: entry.Validator,
			})
		}

		if entry.IsActive && !olderEntry.IsActive {
			entries = append(entries, events.ValidatorActive{
				Validator: entry.Validator,
			})
		}

		if !entry.IsActive && olderEntry.IsActive {
			entries = append(entries, events.ValidatorInactive{
				Validator: entry.Validator,
			})
		}

		if entry.Validator.ConsensusAddressValcons != olderEntry.Validator.ConsensusAddressValcons {
			entries = append(entries, events.ValidatorChangedKey{
				Validator:    entry.Validator,
				OldValidator: olderEntry.Validator,
			})
		}

		if entry.Validator.Moniker != olderEntry.Validator.Moniker {
			entries = append(entries, events.ValidatorChangedMoniker{
				Validator:    entry.Validator,
				OldValidator: olderEntry.Validator,
			})
		}

		if entry.Validator.Commission != olderEntry.Validator.Commission {
			entries = append(entries, events.ValidatorChangedCommission{
				Validator:    entry.Validator,
				OldValidator: olderEntry.Validator,
			})
		}

		isTombstoned := hasNewerSigningInfo && entry.Validator.SigningInfo.Tombstoned
		if isTombstoned || entry.Validator.Jailed || !entry.IsActive {
			continue
		}

		missedBlocksBefore := olderEntry.SignatureInfo.GetNotSigned()
		missedBlocksAfter := entry.SignatureInfo.GetNotSigned()

		beforeGroup, beforeIndex, err := chainConfig.MissedBlocksGroups.GetGroup(missedBlocksBefore)
		if err != nil {
			return nil, err
		}
		afterGroup, afterIndex, err := chainConfig.MissedBlocksGroups.GetGroup(missedBlocksAfter)
		if err != nil {
			return nil, err
		}

		// To fix anomalies, like a validator jumping from 0 to 9500 missed blocks
		if math.Abs(float64(beforeIndex-afterIndex)) > 1 {
			continue
		}

		missedBlocksGroupsEqual := beforeGroup.Start == afterGroup.Start

		if !missedBlocksGroupsEqual && !entry.Validator.Jailed {
			entries = append(entries, events.ValidatorGroupChanged{
				Validator:               entry.Validator,
				MissedBlocksBefore:      missedBlocksBefore,
				MissedBlocksAfter:       missedBlocksAfter,
				MissedBlocksGroupBefore: beforeGroup,
				MissedBlocksGroupAfter:  afterGroup,
			})
		}
	}

	sort.Slice(entries, func(firstIndex, secondIndex int) bool {
		first := entries[firstIndex]
		second := entries[secondIndex]

		// sorting events by their type (e.g. tombstones first, etc, see constants.GetEventNames() for priority
		// if both are EventValidatorGroupChanged, we additionally sort the following:
		// - validators missing blocks are going first, recovering second
		// - if both validators are either skipping or recovering, those skipped more blocks go first

		if first.Type() == constants.EventValidatorGroupChanged && second.Type() == constants.EventValidatorGroupChanged {
			firstConverted, _ := first.(events.ValidatorGroupChanged)
			secondConverted, _ := second.(events.ValidatorGroupChanged)

			// increasing goes first, decreasing goes latest
			if firstConverted.IsIncreasing() != secondConverted.IsIncreasing() {
				return utils.BoolToFloat64(firstConverted.IsIncreasing()) > utils.BoolToFloat64(secondConverted.IsIncreasing())
			}

			return firstConverted.MissedBlocksAfter > secondConverted.MissedBlocksAfter
		}

		firstPriority := slices.Index(constants.GetEventNames(), first.Type())
		secondPriority := slices.Index(constants.GetEventNames(), second.Type())

		return firstPriority < secondPriority
	})

	return &types.Report{Events: entries}, nil
}

type Info struct {
	Height   int64
	Snapshot Snapshot
}
