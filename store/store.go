package store

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/zarbchain/zarb-go/account"
	"github.com/zarbchain/zarb-go/block"
	"github.com/zarbchain/zarb-go/crypto"
	"github.com/zarbchain/zarb-go/crypto/hash"
	"github.com/zarbchain/zarb-go/logger"
	"github.com/zarbchain/zarb-go/tx"
	"github.com/zarbchain/zarb-go/validator"
)

var (
	infoKey         = []byte{0x00}
	blockPrefix     = []byte{0x01}
	blockHashPrefix = []byte{0x03}
	accountPrefix   = []byte{0x05}
	validatorPrefix = []byte{0x07}
	txPrefix        = []byte{0x09}
)

type store struct {
	lk sync.RWMutex

	config         *Config
	db             *leveldb.DB
	batch          *leveldb.Batch
	blockStore     *blockStore
	txStore        *txStore
	accountStore   *accountStore
	validatorStore *validatorStore
}

func NewStore(conf *Config) (Store, error) {
	db, err := leveldb.OpenFile(conf.StorePath(), nil)
	if err != nil {
		return nil, err
	}

	return &store{
		config:         conf,
		db:             db,
		batch:          new(leveldb.Batch),
		blockStore:     newBlockStore(db),
		txStore:        newTxStore(db),
		accountStore:   newAccountStore(db),
		validatorStore: newValidatorStore(db),
	}, nil
}

func (s *store) Close() error {
	return s.db.Close()
}

func (s *store) SaveBlock(height int, block *block.Block) {
	s.lk.Lock()
	defer s.lk.Unlock()
	if err := s.blockStore.saveBlock(s.batch, height, block); err != nil {
		logger.Panic("error on saving block: %v", err)
	}
}

func (s *store) Block(height int) (*block.Block, error) {
	s.lk.Lock()
	defer s.lk.Unlock()

	return s.blockStore.block(height)
}

func (s *store) BlockHeight(hash hash.Hash) (int, error) {
	s.lk.Lock()
	defer s.lk.Unlock()

	return s.blockStore.blockHeight(hash)
}

func (s *store) SaveTransaction(trx *tx.Tx) {
	s.lk.Lock()
	defer s.lk.Unlock()

	if err := s.txStore.saveTx(s.batch, trx); err != nil {
		logger.Panic("error on saving a transaction: %v", err)
	}
}

func (s *store) Transaction(hash hash.Hash) (*tx.Tx, error) {
	s.lk.Lock()
	defer s.lk.Unlock()

	return s.txStore.tx(hash)
}

func (s *store) HasAccount(addr crypto.Address) bool {
	s.lk.Lock()
	defer s.lk.Unlock()

	return s.accountStore.hasAccount(addr)
}

func (s *store) Account(addr crypto.Address) (*account.Account, error) {
	s.lk.Lock()
	defer s.lk.Unlock()

	return s.accountStore.account(addr)
}

func (s *store) TotalAccounts() int {
	s.lk.Lock()
	defer s.lk.Unlock()

	return s.accountStore.total
}

func (s *store) IterateAccounts(consumer func(*account.Account) (stop bool)) {
	s.lk.Lock()
	defer s.lk.Unlock()

	s.accountStore.iterateAccounts(consumer)
}

func (s *store) UpdateAccount(acc *account.Account) {
	s.lk.Lock()
	defer s.lk.Unlock()

	if err := s.accountStore.updateAccount(s.batch, acc); err != nil {
		logger.Panic("error on updating an account: %v", err)
	}
}

func (s *store) HasValidator(addr crypto.Address) bool {
	s.lk.Lock()
	defer s.lk.Unlock()

	return s.validatorStore.hasValidator(addr)
}

func (s *store) Validator(addr crypto.Address) (*validator.Validator, error) {
	s.lk.Lock()
	defer s.lk.Unlock()

	return s.validatorStore.validator(addr)
}

func (s *store) ValidatorByNumber(num int) (*validator.Validator, error) {
	s.lk.Lock()
	defer s.lk.Unlock()

	return s.validatorStore.validatorByNumber(num)
}

func (s *store) TotalValidators() int {
	s.lk.Lock()
	defer s.lk.Unlock()

	return s.validatorStore.total
}

func (s *store) IterateValidators(consumer func(*validator.Validator) (stop bool)) {
	s.lk.Lock()
	defer s.lk.Unlock()

	s.validatorStore.iterateValidators(consumer)
}

func (s *store) UpdateValidator(acc *validator.Validator) {
	s.lk.Lock()
	defer s.lk.Unlock()

	if err := s.validatorStore.updateValidator(s.batch, acc); err != nil {
		logger.Panic("error on updating a validator: %v", err)
	}
}

func (s *store) HasAnyBlock() bool {
	s.lk.Lock()
	defer s.lk.Unlock()

	return s.blockStore.hasAnyBlock()
}
func (s *store) SaveLastInfo(info []byte) {
	s.batch.Put(infoKey, info)
}

func (s *store) RestoreLastInfo() []byte {
	info, _ := tryGet(s.db, infoKey)
	return info
}

func (s *store) WriteBatch() error {
	if err := s.db.Write(s.batch, nil); err != nil {
		return err
	}
	s.batch.Reset()
	return nil
}
