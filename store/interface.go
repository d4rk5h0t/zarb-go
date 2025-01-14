package store

import (
	"github.com/zarbchain/zarb-go/account"
	"github.com/zarbchain/zarb-go/block"
	"github.com/zarbchain/zarb-go/crypto"
	"github.com/zarbchain/zarb-go/crypto/hash"
	"github.com/zarbchain/zarb-go/tx"
	"github.com/zarbchain/zarb-go/validator"
)

type Reader interface {
	Block(height int) (*block.Block, error)
	HasAnyBlock() bool
	BlockHeight(hash hash.Hash) (int, error)
	Transaction(hash hash.Hash) (*tx.Tx, error)
	HasAccount(crypto.Address) bool
	Account(addr crypto.Address) (*account.Account, error)
	TotalAccounts() int
	HasValidator(crypto.Address) bool
	Validator(addr crypto.Address) (*validator.Validator, error)
	ValidatorByNumber(num int) (*validator.Validator, error)
	IterateValidators(consumer func(*validator.Validator) (stop bool))
	IterateAccounts(consumer func(*account.Account) (stop bool))
	TotalValidators() int
	RestoreLastInfo() []byte
}

type Store interface {
	Reader

	UpdateAccount(acc *account.Account)
	UpdateValidator(acc *validator.Validator)
	SaveBlock(height int, block *block.Block)
	SaveTransaction(trx *tx.Tx)
	SaveLastInfo(info []byte)
	WriteBatch() error
	Close() error
}
