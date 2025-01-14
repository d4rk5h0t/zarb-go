package executor

import (
	"github.com/zarbchain/zarb-go/errors"
	"github.com/zarbchain/zarb-go/sandbox"
	"github.com/zarbchain/zarb-go/tx"
	"github.com/zarbchain/zarb-go/tx/payload"
)

type SortitionExecutor struct {
	strict bool
}

func NewSortitionExecutor(strict bool) *SortitionExecutor {
	return &SortitionExecutor{strict: strict}
}

func (e *SortitionExecutor) Execute(trx *tx.Tx, sb sandbox.Sandbox) error {
	pld := trx.Payload().(*payload.SortitionPayload)

	val := sb.Validator(pld.Address)
	if val == nil {
		return errors.Errorf(errors.ErrInvalidTx, "Unable to retrieve validator")
	}
	// Power for parked validators is set to zero
	if val.Power() == 0 {
		return errors.Errorf(errors.ErrInvalidTx, "Validator has no Power to be in committee")
	}
	if sb.CurrentHeight()-val.LastBondingHeight() < sb.BondInterval() {
		return errors.Errorf(errors.ErrInvalidTx, "In bonding period")
	}
	_, hash := sb.FindBlockInfoByStamp(trx.Stamp())
	ok := sb.VerifySortition(hash, pld.Proof, val)
	if !ok {
		return errors.Errorf(errors.ErrInvalidTx, "Sortition proof is invalid")
	}
	if e.strict {
		// A validator might produce more than one sortition transaction before entring into the committee
		// In non-strict mode we don't check the sequence number
		if val.Sequence()+1 != trx.Sequence() {
			return errors.Errorf(errors.ErrInvalidTx, "Invalid sequence. Expected: %v, got: %v", val.Sequence()+1, trx.Sequence())
		}

		if err := sb.EnterCommittee(hash, val.Address()); err != nil {
			return errors.Errorf(errors.ErrInvalidTx, err.Error())
		}
	}

	val.IncSequence()
	val.UpdateLastJoinedHeight(sb.CurrentHeight())

	sb.UpdateValidator(val)

	return nil
}

func (e *SortitionExecutor) Fee() int64 {
	return 0
}
