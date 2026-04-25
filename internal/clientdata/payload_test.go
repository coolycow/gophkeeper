package clientdata

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPayload_roundTripJSON(t *testing.T) {
	p := &Payload{Kind: KindText, Meta: "note", Text: "hello"}
	b, err := p.MarshalJSONBytes()
	require.NoError(t, err)

	out, err := UnmarshalJSONBytes(b)
	require.NoError(t, err)
	require.Equal(t, KindText, out.Kind)
	require.Equal(t, "hello", out.Text)
	require.Equal(t, "note", out.Meta)
}

func TestPayload_Validate_loginPair(t *testing.T) {
	require.Error(t, (&Payload{Kind: KindLoginPair, Login: "u"}).Validate())
	require.NoError(t, (&Payload{Kind: KindLoginPair, Login: "u", Password: "p"}).Validate())
}

func TestSetBinary(t *testing.T) {
	p := &Payload{}
	p.SetBinary([]byte{1, 2, 3})
	require.Equal(t, KindBinary, p.Kind)
	raw, err := p.BinaryBytes()
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3}, raw)
}

func TestPayload_Validate_unknownKind(t *testing.T) {
	err := (&Payload{Kind: "nope", Text: "x"}).Validate()
	require.ErrorIs(t, err, ErrValidation)
}

func TestPayload_Validate_binaryBadBase64(t *testing.T) {
	err := (&Payload{Kind: KindBinary, BinaryBase64: "!!!"}).Validate()
	require.ErrorIs(t, err, ErrValidation)
}

func TestUnmarshal_invalidJSON(t *testing.T) {
	_, err := UnmarshalJSONBytes([]byte("{"))
	require.Error(t, err)
}
