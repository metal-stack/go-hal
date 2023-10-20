package uuid

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func convertUUID(s string) []byte {
	u := uuid.MustParse(s)
	b, _ := u.MarshalBinary()
	return b
}

func TestUuid(t *testing.T) {
	var tests = []struct {
		value  []byte
		expect []byte
	}{
		{
			value:  []byte{0x69, 0xd7, 0x92, 0xd4, 0x3b, 0x2c, 0x11, 0xe4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x8f}, // 69d792d4-3b2c-11e4-0000-00000000008f
			expect: []byte{0xd4, 0x92, 0xd7, 0x69, 0x2c, 0x3b, 0xe4, 0x11, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x8f}, // d492d769-2c3b-e411-0000-00000000008f
		},
		{
			value:  []byte{0x99, 0x34, 0x5, 0x68, 0xf7, 0x75, 0x11, 0xe7, 0x8c, 0x3f, 0x9a, 0x21, 0x4c, 0xf0, 0x93, 0xae}, // 99340568-f775-11e7-8c3f-9a214cf093ae
			expect: []byte{0x68, 0x5, 0x34, 0x99, 0x75, 0xf7, 0xe7, 0x11, 0x8c, 0x3f, 0x9a, 0x21, 0x4c, 0xf0, 0x93, 0xae}, // 68053499-75f7-e711-8c3f-9a214cf093ae
		},
		{
			value:  convertUUID("00ecd471-0771-e911-8000-efbeaddeefbe"),
			expect: convertUUID("71d4ec00-7107-11e9-8000-efbeaddeefbe"),
		},
	}

	for i := range tests {
		tt := tests[i]
		value, err := fromBytes(tt.value)
		if err != nil {
			t.Fatalf("Cannot parse UUID %+v: %s", tt.value, err)
		}

		got, err := value.middleEndianBytes()
		if err != nil {
			t.Fatalf("Cannot convert to middle endian %+v: %s", value, err)
		}

		if !bytes.Equal(got, tt.expect) {
			t.Fatalf("Got %+v, expect %+v", got, tt.expect)
		}
	}
}

func TestUUIDConvert(t *testing.T) {
	bad := "0060c2bd-3089-eb11-8000-7cc255106b08"
	good := "bdc26000-8930-11eb-8000-7cc255106b08"

	result, err := Convert(good)
	require.NoError(t, err)

	assert.Equal(t, good, result)

	result, err = Convert(bad)
	require.NoError(t, err)

	assert.Equal(t, good, result)
}
