package crypto

type PrivateKey interface {
	RawBytes() []byte
	String() string
	MarshalJSON() ([]byte, error)
	MarshalCBOR() ([]byte, error)
	SanityCheck() error
	Sign(msg []byte) Signature
	PublicKey() PublicKey
	EqualsTo(right PrivateKey) bool
}
