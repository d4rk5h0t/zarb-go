package crypto

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddressMarshaling(t *testing.T) {
	addr1 := GenerateTestAddress()
	addr2 := new(Address)
	addr3 := new(Address)
	addr4 := new(Address)

	js, err := json.Marshal(addr1)
	assert.NoError(t, err)
	require.Error(t, addr2.UnmarshalJSON([]byte("bad")))
	require.NoError(t, json.Unmarshal(js, addr2))

	bs, err := addr2.MarshalCBOR()
	assert.NoError(t, err)
	assert.NoError(t, addr3.UnmarshalCBOR(bs))

	txt, err := addr2.MarshalText()
	assert.NoError(t, err)
	assert.NoError(t, addr4.UnmarshalText(txt))

	require.True(t, addr1.EqualsTo(*addr4))
	require.NoError(t, addr1.SanityCheck())
}

func TestAddressFromBytes(t *testing.T) {
	_, err := AddressFromRawBytes(nil)
	assert.Error(t, err)
	addr1 := GenerateTestAddress()
	addr2, err := AddressFromRawBytes(addr1.RawBytes())
	assert.NoError(t, err)
	require.True(t, addr1.EqualsTo(addr2))

	inv, _ := hex.DecodeString("0102")
	_, err = AddressFromRawBytes(inv)
	assert.Error(t, err)
}

func TestAddressFromString(t *testing.T) {
	addr1 := GenerateTestAddress()
	addr2, err := AddressFromString(addr1.String())
	assert.NoError(t, err)
	require.True(t, addr1.EqualsTo(addr2))

	_, err = AddressFromString("inv")
	assert.Error(t, err)
}

func TestMarshalingEmptyAddress(t *testing.T) {
	addr1 := Address{}

	js, err := json.Marshal(addr1)
	assert.NoError(t, err)
	var addr2 Address
	err = json.Unmarshal(js, &addr2)
	assert.NoError(t, err)
	assert.Equal(t, addr1, addr2)

	assert.Error(t, addr2.SanityCheck())

	bs, err := addr1.MarshalCBOR()
	assert.NoError(t, err)
	var addr3 Address
	err = addr3.UnmarshalCBOR(bs)
	assert.NoError(t, err) /// No error
	assert.Equal(t, addr1, addr3)
}

func TestTreasuryAddress(t *testing.T) {
	assert.Equal(t, TreasuryAddress.String(), treasuryAddressString)
	expected, err := AddressFromString(treasuryAddressString)
	assert.NoError(t, err)
	assert.Equal(t, TreasuryAddress.RawBytes(), expected.RawBytes())
}

func TestInvalidBech32(t *testing.T) {
	// ok
	addr, err := AddressFromString("zc17mka0cw484es5whq638xkm89msgzczmrmf7p27")
	assert.NoError(t, err)
	assert.Equal(t, addr.Fingerprint(), "zc17mka0cw48")

	// Invalid hrp
	_, err = AddressFromString("sc17mka0cw484es5whq638xkm89msgzczmr75t2kv")
	assert.Error(t, err)

	// Invalid type
	_, err = AddressFromString("zc27mka0cw484es5whq638xkm89msgzczmrpd86dv")
	assert.Error(t, err)

	// Invalid checksum
	_, err = AddressFromString("zc17mka0cw484es5whq638xkm89msgzczmrwy64dz")
	assert.Error(t, err)
}
