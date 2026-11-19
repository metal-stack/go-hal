package ipmi

import (
	"testing"

	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/stretchr/testify/require"
)

func TestBootFlags(t *testing.T) {
	tests := []struct {
		name   string
		target hal.BootTarget
		vendor api.Vendor
		want   []byte
	}{
		{
			name:   "pxe",
			target: hal.BootTargetPXE,
			vendor: api.VendorLenovo,
			want:   []byte{0xe0, 0x04, 0x00, 0x00, 0x00},
		},
		{
			name:   "disk",
			target: hal.BootTargetDisk,
			vendor: api.VendorLenovo,
			want:   []byte{0xe0, 0x08, 0x00, 0x00, 0x00},
		},
		{
			name:   "bios",
			target: hal.BootTargetBIOS,
			vendor: api.VendorLenovo,
			want:   []byte{0xa0, 0x18, 0x00, 0x00, 0x00},
		},
		{
			name:   "disk supermicro",
			target: hal.BootTargetDisk,
			vendor: api.VendorSupermicro,
			want:   []byte{0xe0, 0x24, 0x00, 0x00, 0x00},
		},
		{
			name:   "disk vagrant",
			target: hal.BootTargetDisk,
			vendor: api.VendorVagrant,
			want:   []byte{0xe0, 0x08, 0x00, 0x00, 0x00},
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, BootFlags(tt.target, tt.vendor).Pack())
		})
	}
}

func TestLanConfigString(t *testing.T) {
	c := LanConfig{IP: "10.0.0.1", Mac: "aa:bb:cc:dd:ee:ff"}
	require.Equal(t, "ip: 10.0.0.1 mac:aa:bb:cc:dd:ee:ff", c.String())
}
