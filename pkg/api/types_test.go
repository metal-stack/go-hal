package api

import (
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
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.String(); got != tt.want {
				t.Errorf("Vendor.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGuessVendor(t *testing.T) {
	tests := []struct {
		name   string
		vendor string
		want   Vendor
	}{
		{name: "smc", vendor: "Supermicro", want: VendorSupermicro},
		{name: "smc with space", vendor: " Supermicro ", want: VendorSupermicro},
		{name: "lenovo", vendor: "Lenovo", want: VendorLenovo},
		{name: "lenovolower", vendor: "lenovo", want: VendorLenovo},
		{name: "empty", vendor: "", want: VendorUnknown},
		{name: "unknown", vendor: "unknown", want: VendorUnknown},
		{name: "vagrant", vendor: "vagrant", want: VendorVagrant},
		{name: "dell", vendor: "Dell", want: VendorDell},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if got := GuessVendor(tt.vendor); got != tt.want {
				t.Errorf("GuessVendor() = %v, want %v", got, tt.want)
			}
		})
	}
}
