package detect

import (
	"reflect"
	"testing"
)

func TestVendor_String(t *testing.T) {
	tests := []struct {
		name string
		v    Vendor
		want string
	}{
		{name: "smc", v: VendorSupermicro, want: "Supermicro"},
		{name: "dell", v: VendorDell, want: "Dell"},
		{name: "lenovo", v: VendorLenovo, want: "Lenovo"},
		{name: "unknown", v: VendorUnknown, want: "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.String(); got != tt.want {
				t.Errorf("Vendor.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectInBand(t *testing.T) {
	tests := []struct {
		name    string
		want    *Board
		wantErr bool
	}{
		{
			name:    "simple",
			want:    &Board{Vendor: VendorLenovo, Name: "20F9S0G500"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InBand()
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectInBand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DetectInBand() = %v, want %v", got, tt.want)
			}
		})
	}
}
