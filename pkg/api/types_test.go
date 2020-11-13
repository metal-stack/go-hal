package api

import "testing"

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
