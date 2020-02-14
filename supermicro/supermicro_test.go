package hal

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func Test_inBand_UUID(t *testing.T) {
	tests := []struct {
		name    string
		s       *inBand
		want    uuid.UUID
		wantErr bool
	}{
		{name: "not implemented", s: &inBand{}, want: uuid.UUID{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &inBand{}
			got, err := s.UUID()
			if (err != nil) != tt.wantErr {
				t.Errorf("inBand.UUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("inBand.UUID() = %v, want %v", got, tt.want)
			}
		})
	}
}
