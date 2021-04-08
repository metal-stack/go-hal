package uuid

import (
	"bytes"
	"reflect"
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

	for i := range tests {
		tt := tests[i]
		value, err := FromBytes(tt.value)
		if err != nil {
			t.Fatalf("Cannot parse UUID %+v: %s", tt.value, err)
		}

		got, err := value.MiddleEndianBytes()
		if err != nil {
			t.Fatalf("Cannot convert to middle endian %+v: %s", value, err)
		}

		if !bytes.Equal(got, tt.expect) {
			t.Fatalf("Got %+v, expect %+v", got, tt.expect)
		}
	}
}

func TestFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "creates uuid",
			input:   "00ecd471-0771-e911-8000-efbeaddeefbe",
			wantErr: false,
		},
		{
			name:    "creates uuid",
			input:   "11ecd471-2771-e911-8333-efbeaddeefbe",
			wantErr: false,
		},
		{
			name:    "creates uuid",
			input:   "11ecd471-2771-e911-8333-0000addeefbe",
			wantErr: false,
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.String(), tt.input) {
				t.Errorf("FromString() = %v, want %v", got.String(), tt.input)
			}
		})
	}
}

func TestUuid_ToMiddleEndian(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "convert to mixed endian",
			input:   "00ecd471-0771-e911-8000-efbeaddeefbe",
			want:    "71d4ec00-7107-11e9-8000-efbeaddeefbe",
			wantErr: false,
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			u, err := FromString(tt.input)
			if err != nil {
				t.Errorf("Uuid.FormatString() error = %v", err)
				return
			}
			got, err := u.ToMiddleEndian()
			if (err != nil) != tt.wantErr {
				t.Errorf("Uuid.ToMiddleEndian() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.String(), tt.want) {
				t.Errorf("Uuid.ToMiddleEndian() = %v, want %v", got.String(), tt.want)
			}
		})
	}
}
