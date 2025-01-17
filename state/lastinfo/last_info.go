package lastinfo

import (
	"fmt"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/zarbchain/zarb-go/block"
	"github.com/zarbchain/zarb-go/committee"
	"github.com/zarbchain/zarb-go/crypto"
	"github.com/zarbchain/zarb-go/crypto/hash"
	"github.com/zarbchain/zarb-go/logger"
	"github.com/zarbchain/zarb-go/sortition"
	"github.com/zarbchain/zarb-go/store"
	"github.com/zarbchain/zarb-go/tx/payload"
	"github.com/zarbchain/zarb-go/validator"
)

type lastInfoData struct {
	LastBlockHeight int                `cbor:"1,keyasint"`
	LastCertificate *block.Certificate `cbor:"2,keyasint"`
}

type LastInfo struct {
	lk sync.RWMutex

	store             store.Store
	lastBlockHeight   int
	lastSortitionSeed sortition.VerifiableSeed
	lastBlockHash     hash.Hash
	lastCertificate   *block.Certificate
	lastBlockTime     time.Time
}

func NewLastInfo(store store.Store) *LastInfo {
	return &LastInfo{store: store}
}

func (li *LastInfo) SortitionSeed() sortition.VerifiableSeed {
	li.lk.RLock()
	defer li.lk.RUnlock()

	return li.lastSortitionSeed
}

func (li *LastInfo) BlockHeight() int {
	li.lk.RLock()
	defer li.lk.RUnlock()

	return li.lastBlockHeight
}

func (li *LastInfo) BlockHash() hash.Hash {
	li.lk.RLock()
	defer li.lk.RUnlock()

	return li.lastBlockHash
}

func (li *LastInfo) Certificate() *block.Certificate {
	li.lk.RLock()
	defer li.lk.RUnlock()

	return li.lastCertificate
}

func (li *LastInfo) BlockTime() time.Time {
	li.lk.RLock()
	defer li.lk.RUnlock()

	return li.lastBlockTime
}

func (li *LastInfo) SetSortitionSeed(lastSortitionSeed sortition.VerifiableSeed) {
	li.lk.Lock()
	defer li.lk.Unlock()

	li.lastSortitionSeed = lastSortitionSeed
}

func (li *LastInfo) SetBlockHeight(lastBlockHeight int) {
	li.lk.Lock()
	defer li.lk.Unlock()

	li.lastBlockHeight = lastBlockHeight
}

func (li *LastInfo) SetBlockHash(lastBlockHash hash.Hash) {
	li.lk.Lock()
	defer li.lk.Unlock()

	li.lastBlockHash = lastBlockHash
}

func (li *LastInfo) SetCertificate(lastCertificate *block.Certificate) {
	li.lk.Lock()
	defer li.lk.Unlock()

	li.lastCertificate = lastCertificate
}

func (li *LastInfo) SetBlockTime(lastBlockTime time.Time) {
	li.lk.Lock()
	defer li.lk.Unlock()

	li.lastBlockTime = lastBlockTime
}

func (li *LastInfo) SaveLastInfo() {
	lid := lastInfoData{
		LastBlockHeight: li.lastBlockHeight,
		LastCertificate: li.lastCertificate,
	}

	bs, _ := cbor.Marshal(&lid)
	li.store.SaveLastInfo(bs)
}

func (li *LastInfo) RestoreLastInfo(committeeSize int, srt *sortition.Sortition) (*committee.Committee, error) {
	bs := li.store.RestoreLastInfo()
	lid := new(lastInfoData)
	err := cbor.Unmarshal(bs, lid)
	if err != nil {
		return nil, err
	}

	logger.Debug("try to restore last state info", "height", lid.LastBlockHeight)

	b, err := li.store.Block(lid.LastBlockHeight)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve block %v: %v", lid.LastBlockHeight, err)
	}

	li.lastBlockHeight = lid.LastBlockHeight
	li.lastCertificate = lid.LastCertificate
	li.lastSortitionSeed = b.Header().SortitionSeed()
	li.lastBlockHash = b.Hash()
	li.lastBlockTime = b.Header().Time()

	cmt, err := li.restoreCommittee(committeeSize)
	if err != nil {
		return nil, err
	}

	err = li.restoreSortition(srt, cmt)
	if err != nil {
		return nil, err
	}

	return cmt, nil
}

func (li *LastInfo) restoreCommittee(committeeSize int) (*committee.Committee, error) {
	b, _ := li.store.Block(li.lastBlockHeight)

	joinedVals := make([]*validator.Validator, 0)
	for _, id := range b.TxIDs().IDs() {
		trx, err := li.store.Transaction(id)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve transaction %s: %v", id, err)
		}
		// If there is any sortition transaction in last block,
		// we should update last committee
		if trx.IsSortitionTx() {
			pld := trx.Payload().(*payload.SortitionPayload)
			val, err := li.store.Validator(pld.Address)
			if err != nil {
				return nil, fmt.Errorf("unable to retrieve validator %s: %v", pld.Address, err)
			}
			joinedVals = append(joinedVals, val)
		}
	}

	proposerIndex := 0
	curCommitteeSize := len(li.lastCertificate.Committers())
	vals := make([]*validator.Validator, len(li.lastCertificate.Committers()))
	for i, num := range li.lastCertificate.Committers() {
		val, err := li.store.ValidatorByNumber(num)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve committee member %v: %v", num, err)
		}
		if b.Header().ProposerAddress() == val.Address() {
			proposerIndex = i
		}
		vals[i] = val
	}

	// First we create previous committee, the we update it to get the latest committee.
	proposerIndex = (proposerIndex + curCommitteeSize - (li.lastCertificate.Round() % curCommitteeSize)) % curCommitteeSize
	committee, err := committee.NewCommittee(vals, committeeSize, vals[proposerIndex].Address())
	if err != nil {
		return nil, fmt.Errorf("unable to create last committee: %v", err)
	}

	err = committee.Update(li.lastCertificate.Round(), joinedVals)
	if err != nil {
		return nil, fmt.Errorf("unable to update last committee: %v", err)
	}

	return committee, nil
}

func (li *LastInfo) restoreSortition(srt *sortition.Sortition, cmt *committee.Committee) error {
	type sortitionParam struct {
		blockHash hash.Hash
		seed      sortition.VerifiableSeed
		poolStake int64
	}

	totalStake := int64(0)
	li.store.IterateValidators(func(v *validator.Validator) (stop bool) {
		totalStake += v.Stake()
		return false
	})

	params := []sortitionParam{}

	// Read last seven blocks
	start := li.lastBlockHeight - 7
	if start < 0 {
		start = 0
	}

	stakeChanged := make(map[crypto.Address]int64)
	cert := li.lastCertificate
	curCommitters := cmt.Committers()
	for h := li.lastBlockHeight; h > start; h-- {
		b, err := li.store.Block(h)
		if err != nil {
			return fmt.Errorf("unable to retrieve block %v: %v", h, err)
		}

		committeeStake := int64(0)
		for _, num := range curCommitters {
			val, err := li.store.ValidatorByNumber(num)
			if err != nil {
				return fmt.Errorf("unable to retrieve committee member %v: %v", num, err)
			}

			committeeStake += val.Stake()
			changed, ok := stakeChanged[val.Address()]
			if ok {
				committeeStake += changed
			}
		}

		param := sortitionParam{
			blockHash: b.Hash(),
			seed:      b.Header().SortitionSeed(),
			poolStake: totalStake - committeeStake,
		}
		params = append(params, param)

		for _, id := range b.TxIDs().IDs() {
			trx, err := li.store.Transaction(id)
			if err != nil {
				return fmt.Errorf("unable to retrieve transaction %s: %v", id, err)
			}
			if trx.IsBondTx() {
				pld := trx.Payload().(*payload.BondPayload)
				totalStake -= pld.Stake
				stakeChanged[pld.PublicKey.Address()] = stakeChanged[pld.PublicKey.Address()] - pld.Stake
			}
		}
		curCommitters = cert.Committers()
		cert = b.PrevCertificate()
	}

	for i := len(params) - 1; i >= 0; i-- {
		p := params[i]
		srt.SetParams(p.blockHash, p.seed, p.poolStake)
	}

	return nil
}
