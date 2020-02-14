package hal

import (
	"fmt"

	"github.com/google/uuid"
	hal "github.com/metal-stack/go-hal"
)

type (
	inBand struct {
	}
	outBand struct {
	}
)

var (
	// ErrorNotImplemented for all funcs which are not implemented yet
	ErrorNotImplemented = fmt.Errorf("not implemented yet")
)

// InBand create a inband connection to a supermicro server.
func InBand() (hal.InBand, error) {
	return &inBand{}, nil
}

// OutBand create a outband connection to a supermicro server.
func OutBand() (hal.OutBand, error) {
	return &outBand{}, nil
}

// InBand

func (s *inBand) UUID() (uuid.UUID, error) {
	return uuid.UUID{}, ErrorNotImplemented
}
func (s *inBand) PowerOff() error {
	return ErrorNotImplemented
}
func (s *inBand) PowerReset() error {
	return ErrorNotImplemented
}
func (s *inBand) PowerCycle() error {
	return ErrorNotImplemented
}
func (s *inBand) BootFrom(hal.BootTarget) error {
	return ErrorNotImplemented
}
func (s *inBand) Firmware() (hal.FirmwareMode, error) {
	return hal.FirmwareModeUnknown, ErrorNotImplemented
}
func (s *inBand) SetFirmware(hal.FirmwareMode) error {
	return ErrorNotImplemented
}

// OutBand

func (s *outBand) UUID() (uuid.UUID, error) {
	return uuid.UUID{}, ErrorNotImplemented
}
func (s *outBand) PowerState() (hal.PowerState, error) {
	return hal.PowerUnknownState, ErrorNotImplemented
}
func (s *outBand) PowerOn() error {
	return ErrorNotImplemented
}
func (s *outBand) PowerOff() error {
	return ErrorNotImplemented
}

func (s *outBand) PowerReset() error {
	return ErrorNotImplemented
}

func (s *outBand) PowerCycle() error {
	return ErrorNotImplemented
}

func (s *outBand) IdentifyLEDState(hal.IdentifyLEDState) error {
	return ErrorNotImplemented
}

func (s *outBand) IdentifyLEDOn() error {
	return ErrorNotImplemented
}

func (s *outBand) IdentifyLEDOff() error {
	return ErrorNotImplemented
}

func (s *outBand) BootFrom(hal.BootTarget) error {
	return ErrorNotImplemented
}
