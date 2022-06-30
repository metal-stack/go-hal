package dmi

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestDMI_BoardInfo(t *testing.T) {
	tests := []struct {
		name   string
		mockFn func(fs afero.Fs)
		want   *api.Board
	}{
		{
			name: "reading board info",
			mockFn: func(fs afero.Fs) {
				require.NoError(t, afero.WriteFile(fs, boardVendor, []byte("vendor"), 0644))
				require.NoError(t, afero.WriteFile(fs, boardName, []byte("name"), 0644))
				require.NoError(t, afero.WriteFile(fs, boardSerial, []byte("serial"), 0644))
				require.NoError(t, afero.WriteFile(fs, productSerial, []byte("product serial"), 0644))
				require.NoError(t, afero.WriteFile(fs, biosVersion, []byte("bios"), 0644))
			},
			want: &api.Board{
				VendorString: "vendor",
				Model:        "name",
				PartNumber:   "product serial",
				SerialNumber: "serial",
				BiosVersion:  "bios",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.mockFn != nil {
				tt.mockFn(fs)
			}

			d := &DMI{
				log: zaptest.NewLogger(t).Sugar(),
				fs:  fs,
			}

			got, err := d.BoardInfo()
			assert.NoError(t, err)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("BoardInfo() assertion failed (+got -want):\n %v", diff)
			}
		})
	}
}
