package dmi

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestMachineUUID(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(fs afero.Fs)
		want    string
		wantErr error
	}{
		{
			name:    "no file present",
			want:    "",
			wantErr: ErrNoUUIDFound,
		},
		{
			name: "reading from " + productUUID,
			mockFn: func(fs afero.Fs) {
				require.NoError(t, afero.WriteFile(fs, productUUID, []byte("4c4c4544-0042-4810-8056-b4c04f395332"), 0644))
			},
			want: "4c4c4544-0042-4810-8056-b4c04f395332",
		},
		{
			name: "reading from " + productSerial,
			mockFn: func(fs afero.Fs) {
				require.NoError(t, afero.WriteFile(fs, productSerial, []byte("4c4c4544-0042-4810-8056-b4c04f395332"), 0644))
			},
			want: "4c4c4544-0042-4810-8056-b4c04f395332",
		},
		{
			name: "reading invalid serial from " + productSerial,
			mockFn: func(fs afero.Fs) {
				err := afero.WriteFile(fs, productSerial, []byte("HDZ8P73"), 0644)
				require.NoError(t, err)
			},
			want:    "",
			wantErr: ErrNoUUIDFound,
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.mockFn != nil {
				tt.mockFn(fs)
			}

			d := &DMI{
				log: zaptest.NewLogger(t).Sugar(),
				fs:  fs,
			}

			got, err := d.MachineUUID()

			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("MachineUUID() assertion failed (+got -want):\n %v", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("MachineUUID() assertion failed (+got -want):\n %v", diff)
			}
		})
	}
}
