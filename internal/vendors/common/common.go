package common

import (
	"github.com/google/uuid"
	hal "github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/dmi"
	"github.com/metal-stack/go-hal/internal/kernel"
	"github.com/metal-stack/go-hal/internal/redfish"
)

type Common struct {
	*redfish.APIClient
}

func New(r *redfish.APIClient) *Common {
	return &Common{
		APIClient: r,
	}
}

func (c *Common) UUID() (*uuid.UUID, error) {
	u, err := dmi.MachineUUID()
	if err != nil {
		return nil, err
	}
	us, err := uuid.Parse(u)
	if err != nil {
		return nil, err
	}
	return &us, nil
}

func (c *Common) Firmware() (hal.FirmwareMode, error) {
	var firmware hal.FirmwareMode
	switch kernel.Firmware() {
	case kernel.BIOS:
		firmware = hal.FirmwareModeLegacy
	case kernel.EFI:
		firmware = hal.FirmwareModeUEFI
	default:
		firmware = hal.FirmwareModeUnknown
	}
	return firmware, nil
}
