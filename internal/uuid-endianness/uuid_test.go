package uuid

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
)

func ConvertUUID(s string) []byte {
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
			value:  ConvertUUID("00ecd471-0771-e911-8000-efbeaddeefbe"),
			expect: ConvertUUID("71d4ec00-7107-11e9-8000-efbeaddeefbe"),
		},
	}

	for _, test := range tests {
		value, err := FromBytes(test.value)
		if err != nil {
			t.Fatalf("Cannot parse UUID %+v: %s", test.value, err)
		}

		got, err := value.MiddleEndianBytes()
		if err != nil {
			t.Fatalf("Cannot convert to middle endian %+v: %s", value, err)
		}

		if !bytes.Equal(got, test.expect) {
			t.Fatalf("Got %+v, expect %+v", got, test.expect)
		}
	}
}
