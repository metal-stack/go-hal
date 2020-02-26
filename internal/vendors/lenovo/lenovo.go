package lenovo

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/dmi"
	"github.com/metal-stack/go-hal/internal/kernel"
)

type (
	inBand struct {
	}
	outBand struct {
	}
)

var (
	// errorNotImplemented for all funcs which are not implemented yet
	errorNotImplemented = fmt.Errorf("not implemented yet")
)

// InBand create a inband connection to a supermicro server.
func InBand() (hal.InBand, error) {
	return &inBand{}, nil
}

// OutBand create a outband connection to a supermicro server.
func OutBand(ip, user, password *string) (hal.OutBand, error) {
	return &outBand{}, nil
}

// InBand

func (s *inBand) UUID() (*uuid.UUID, error) {
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
func (s *inBand) PowerOff() error {
	return errorNotImplemented
}
func (s *inBand) PowerReset() error {
	return errorNotImplemented
}
func (s *inBand) PowerCycle() error {
	return errorNotImplemented
}
func (s *inBand) BootFrom(hal.BootTarget) error {
	return errorNotImplemented
}
func (s *inBand) Firmware() (hal.FirmwareMode, error) {
	var firmware hal.FirmwareMode
	switch kernel.Firmware() {
	case "bios":
		firmware = hal.FirmwareModeLegacy
	case "efi":
		firmware = hal.FirmwareModeUEFI
	default:
		firmware = hal.FirmwareModeUnknown
	}
	return firmware, nil
}
func (s *inBand) SetFirmware(hal.FirmwareMode) error {
	return errorNotImplemented
}

// OutBand

func (s *outBand) UUID() (*uuid.UUID, error) {
	return nil, errorNotImplemented
}
func (s *outBand) PowerState() (hal.PowerState, error) {
	return hal.PowerUnknownState, errorNotImplemented
}
func (s *outBand) PowerOn() error {
	return errorNotImplemented
}
func (s *outBand) PowerOff() error {
	return errorNotImplemented
}

func (s *outBand) PowerReset() error {
	return errorNotImplemented
}

func (s *outBand) PowerCycle() error {
	return errorNotImplemented
}

func (s *outBand) IdentifyLEDState(hal.IdentifyLEDState) error {
	return errorNotImplemented
}

func (s *outBand) IdentifyLEDOn() error {
	return errorNotImplemented
}

func (s *outBand) IdentifyLEDOff() error {
	return errorNotImplemented
}

func (s *outBand) BootFrom(hal.BootTarget) error {
	return errorNotImplemented
}
