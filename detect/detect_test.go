package detect

import (
	"reflect"
	"testing"

	"github.com/metal-stack/go-hal/pkg/api"
)

func _TestDetectInBand(t *testing.T) {
	tests := []struct {
		name    string
		want    api.Vendor
		wantErr bool
	}{
		{
			name:    "simple",
			want:    api.VendorLenovo,
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
