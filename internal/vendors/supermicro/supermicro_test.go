package supermicro

import (
	"testing"

	"github.com/google/uuid"
	"github.com/metal-stack/go-hal"
	uuidendian "github.com/metal-stack/go-hal/internal/uuid-endianness"
	"github.com/stretchr/testify/require"
)

func Test_UUID(t *testing.T) {
	// given
	u := "f6157800-70e3-11e9-8000-efbeaddeefbe"

	uid, err := uuidendian.FromString(u)

	// then
	require.NoError(t, err)
	require.NotNil(t, uid)
	require.Equal(t, "f6157800-70e3-11e9-8000-efbeaddeefbe", uid.String())

	// when
	id, err := UUID(u)

	// then
	require.NoError(t, err)
	require.NotNil(t, id)
	require.Equal(t, "007815f6-e370-e911-8000-efbeaddeefbe", id.String())
}

func UUID(u string) (*uuid.UUID, error) {
	raw, err := uuidendian.FromString(u)
	if err != nil {
		return nil, err
	}

	mixed, err := raw.ToMiddleEndian()
	if err != nil {
		return nil, err
	}

	us, err := uuid.Parse(mixed.String())
	if err != nil {
		return nil, err
	}
	return &us, nil
}

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
