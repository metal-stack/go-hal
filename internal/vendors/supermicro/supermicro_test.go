package supermicro

import (
	"testing"
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
		{name: "not implemented", fields: fields{sum: &sum{sum: "/bin/true"}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &inBand{
				sum: tt.fields.sum,
			}
			if err := s.PowerOff(); (err != nil) != tt.wantErr {
				t.Errorf("inBand.PowerOff() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
