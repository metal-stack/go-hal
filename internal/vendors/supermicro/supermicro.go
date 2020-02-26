package supermicro

import (
	"fmt"

	"github.com/google/uuid"
	hal "github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/dmi"
	"github.com/metal-stack/go-hal/internal/kernel"
)

type (
	inBand struct {
		sum *sum
	}
	outBand struct {
		sum *sum
	}
)

var (
	// errorNotImplemented for all funcs which are not implemented yet
	errorNotImplemented = fmt.Errorf("not implemented yet")
)

// InBand create a inband connection to a supermicro server.
func InBand(sumBin string) (hal.InBand, error) {
	s, err := newSum(sumBin, false, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return &inBand{
		sum: s,
	}, nil
}

// OutBand create a outband connection to a supermicro server.
func OutBand(sumBin string, remote bool, ip, user, password *string) (hal.OutBand, error) {
	s, err := newSum(sumBin, remote, ip, user, password)
	if err != nil {
		return nil, err
	}
	return &outBand{
		sum: s,
	}, nil
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
	u, err := s.sum.uuidRemote()
	if err != nil {
		return nil, err
	}
	us, err := uuid.Parse(u)
	if err != nil {
		return nil, err
	}
	return &us, errorNotImplemented
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
