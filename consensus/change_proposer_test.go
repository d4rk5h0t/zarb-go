package consensus

import (
	"testing"
	"time"

	"github.com/zarbchain/zarb-go/consensus/vote"
	"github.com/zarbchain/zarb-go/crypto/hash"
)

func TestChangeProposer(t *testing.T) {
	setup(t)

	tConsP.config.ChangeProposerTimeout = 100 * time.Millisecond
	testEnterNewHeight(tConsP)

	shouldPublishVote(t, tConsP, vote.VoteTypeChangeProposer, hash.UndefHash)
}

func TestGotoNewRound(t *testing.T) {
	setup(t)

	commitBlockForAllStates(t)

	tConsP.config.ChangeProposerTimeout = 100 * time.Millisecond
	testEnterNewHeight(tConsP)

	shouldPublishVote(t, tConsP, vote.VoteTypeChangeProposer, hash.UndefHash)
	testAddVote(tConsP, vote.VoteTypeChangeProposer, 2, 0, hash.UndefHash, tIndexX)
	testAddVote(tConsP, vote.VoteTypeChangeProposer, 2, 0, hash.UndefHash, tIndexY)

	checkHeightRound(t, tConsP, 2, 1)
}
