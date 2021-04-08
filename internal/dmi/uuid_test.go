package dmi

import (
	"testing"
)

func TestMachineUUID(t *testing.T) {
	readFileFunc := func(filename string) ([]byte, error) {
		return []byte("4C4C4544-0042-4810-8056-B4C04F395332"), nil
	}

	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "TestMachineUUID Test 1",
			want:    "4C4C4544-0042-4810-8056-B4C04F395332",
			wantErr: false,
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if got, err := machineUUID(readFileFunc); got != tt.want {
				if err == nil && tt.wantErr {
					t.Errorf("MachineUUID() = %v, want %v", err, tt.wantErr)
				}
				t.Errorf("MachineUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}
