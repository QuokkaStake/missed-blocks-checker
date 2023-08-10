package snapshot

import (
	"golang.org/x/exp/slices"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/events"
	"main/pkg/report"
	"main/pkg/types"
	"math"
	"sort"
)

type Entry struct {
	Validator     *types.Validator
	SignatureInfo types.SignatureInto
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

type Snapshot struct {
	Entries Entries
}

func (snapshot *Snapshot) GetReport(
	olderSnapshot Snapshot,
	chainConfig *config.ChainConfig,
) (*report.Report, error) {
	var entries []report.Entry

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

		if entry.Validator.Jailed && !olderEntry.Validator.Jailed && olderEntry.Validator.Active() {
			entries = append(entries, events.ValidatorJailed{
				Validator: entry.Validator,
			})
		}

		if !entry.Validator.Jailed && olderEntry.Validator.Jailed {
			entries = append(entries, events.ValidatorUnjailed{
				Validator: entry.Validator,
			})
		}

		if entry.Validator.Active() && !olderEntry.Validator.Active() {
			entries = append(entries, events.ValidatorActive{
				Validator: entry.Validator,
			})
		}

		if !entry.Validator.Active() && olderEntry.Validator.Active() {
			entries = append(entries, events.ValidatorInactive{
				Validator: entry.Validator,
			})
		}

		isTombstoned := hasNewerSigningInfo && entry.Validator.SigningInfo.Tombstoned
		if isTombstoned || entry.Validator.Jailed || !entry.Validator.Active() {
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

	sort.Slice(entries, func(first, second int) bool {
		firstPriority := slices.Index(constants.GetEventNames(), entries[first].Type())
		secondPriority := slices.Index(constants.GetEventNames(), entries[second].Type())

		return firstPriority < secondPriority
	})

	return &report.Report{Entries: entries}, nil
}

type Info struct {
	Height   int64
	Snapshot Snapshot
}
