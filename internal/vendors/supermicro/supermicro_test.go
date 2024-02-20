package supermicro

import (
	"testing"

	"github.com/metal-stack/go-hal"
)

func Test_inBand_PowerOff(t *testing.T) {
	type fields struct {
		sum *sum
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "not implemented", fields: fields{sum: &sum{binary: "/bin/true"}}, wantErr: true},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			s := &inBand{
				sum: tt.fields.sum,
			}
			if err := s.SetFirmware(hal.FirmwareModeUEFI); (err != nil) != tt.wantErr {
				t.Errorf("inBand.PowerOff() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
