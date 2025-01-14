package message

import (
	"fmt"

	"github.com/zarbchain/zarb-go/crypto/hash"
	"github.com/zarbchain/zarb-go/errors"
)

type HeartBeatMessage struct {
	Height        int       `cbor:"1,keyasint"`
	Round         int       `cbor:"2,keyasint"`
	PrevBlockHash hash.Hash `cbor:"3,keyasint"`
}

func NewHeartBeatMessage(h, r int, hash hash.Hash) *HeartBeatMessage {
	return &HeartBeatMessage{
		Height:        h,
		Round:         r,
		PrevBlockHash: hash,
	}
}

func (m *HeartBeatMessage) SanityCheck() error {
	if m.Height <= 0 {
		return errors.Errorf(errors.ErrInvalidMessage, "invalid height")
	}
	if m.Round < 0 {
		return errors.Errorf(errors.ErrInvalidMessage, "invalid round")
	}
	return nil
}

func (m *HeartBeatMessage) Type() Type {
	return MessageTypeHeartBeat
}

func (m *HeartBeatMessage) Fingerprint() string {
	return fmt.Sprintf("{%d/%d}", m.Height, m.Round)
}
