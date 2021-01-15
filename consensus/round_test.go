package consensus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zarbchain/zarb-go/block"

	"github.com/zarbchain/zarb-go/consensus/hrs"
	"github.com/zarbchain/zarb-go/crypto"
	"github.com/zarbchain/zarb-go/vote"
)

func TestInvalidBlock(t *testing.T) {
	setup(t)

	a := tSigners[tIndexB].Address()
	invBlock, _ := block.GenerateTestBlock(&a, nil)
	p := vote.NewProposal(1, 2, *invBlock)
	tSigners[tIndexB].SignMsg(p)

	tConsY.enterNewHeight()
	tConsY.enterNewRound(2)
	tConsY.SetProposal(p)
	assert.Nil(t, tConsY.LastProposal())

}
func TestNewRound(t *testing.T) {
	setup(t)

	tConsP.MoveToNewHeight()
	checkHRSWait(t, tConsP, 1, 0, hrs.StepTypePropose)

	//
	// 1- Move to round 0
	// 2- PreCommits  for round 0 => missed
	// 3- PreCommits  for round 1 => missed
	// 4- PreCommits  for round 2 => received
	// 5- Moved to round 3
	// 6- PreCommits  for round 0 => received
	// 7- Should ignore moving to round 1

	testAddVote(t, tConsP, vote.VoteTypePrecommit, 1, 2, crypto.UndefHash, tIndexX, false)
	testAddVote(t, tConsP, vote.VoteTypePrecommit, 1, 2, crypto.UndefHash, tIndexY, false)
	testAddVote(t, tConsP, vote.VoteTypePrecommit, 1, 2, crypto.UndefHash, tIndexB, false)

	checkHRSWait(t, tConsP, 1, 3, hrs.StepTypePrepare)

	testAddVote(t, tConsP, vote.VoteTypePrecommit, 1, 0, crypto.UndefHash, tIndexX, false)
	testAddVote(t, tConsP, vote.VoteTypePrecommit, 1, 0, crypto.UndefHash, tIndexY, false)
	testAddVote(t, tConsP, vote.VoteTypePrecommit, 1, 0, crypto.UndefHash, tIndexB, false)

	checkHRS(t, tConsP, 1, 3, hrs.StepTypePrepare)
}

func TestConsensusGotoNextRound(t *testing.T) {
	setup(t)

	tConsY.enterNewHeight()

	// Validator_1 is offline
	testAddVote(t, tConsY, vote.VoteTypePrepare, 1, 0, crypto.UndefHash, tIndexB, false)
	testAddVote(t, tConsY, vote.VoteTypePrepare, 1, 0, crypto.UndefHash, tIndexP, false)
	checkHRSWait(t, tConsY, 1, 0, hrs.StepTypePrecommit)

	testAddVote(t, tConsY, vote.VoteTypePrecommit, 1, 0, crypto.UndefHash, tIndexB, false)
	testAddVote(t, tConsY, vote.VoteTypePrecommit, 1, 0, crypto.UndefHash, tIndexP, false)
	checkHRSWait(t, tConsY, 1, 1, hrs.StepTypePrepare)

}

func TestConsensusGotoNextRound2(t *testing.T) {
	setup(t)

	tConsY.enterNewHeight()

	// Validator_1 is offline
	testAddVote(t, tConsY, vote.VoteTypePrepare, 1, 0, crypto.GenerateTestHash(), tIndexB, false)
	testAddVote(t, tConsY, vote.VoteTypePrepare, 1, 0, crypto.GenerateTestHash(), tIndexP, false)
	checkHRSWait(t, tConsY, 1, 0, hrs.StepTypePrecommit)

	testAddVote(t, tConsY, vote.VoteTypePrecommit, 1, 0, crypto.UndefHash, tIndexB, false)
	testAddVote(t, tConsY, vote.VoteTypePrecommit, 1, 0, crypto.UndefHash, tIndexP, false)
	checkHRSWait(t, tConsY, 1, 1, hrs.StepTypePrepare)
}