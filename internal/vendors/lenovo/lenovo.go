package lenovo

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/internal/vendors/common"
)

type (
	inBand struct {
		common *common.Common
	}
	outBand struct {
		redfish *redfish.APIClient
		common  *common.Common
	}
)

var (
	// errorNotImplemented for all funcs which are not implemented yet
	errorNotImplemented = fmt.Errorf("not implemented yet")
)

// InBand create a inband connection to a supermicro server.
func InBand() (hal.InBand, error) {
	return &inBand{
		common: common.New(nil),
	}, nil
}

// OutBand create a outband connection to a supermicro server.
func OutBand(ip, user, password *string) (hal.OutBand, error) {
	r, err := redfish.New("https://"+*ip, *user, *password, true)
	if err != nil {
		return nil, err
	}
	return &outBand{
		redfish: r,
		common:  common.New(r),
	}, nil
}

// InBand

func (i *inBand) UUID() (*uuid.UUID, error) {
	return i.common.UUID()
}
func (i *inBand) PowerOff() error {
	return errorNotImplemented
}
func (i *inBand) PowerReset() error {
	return errorNotImplemented
}
func (i *inBand) PowerCycle() error {
	return errorNotImplemented
}
func (i *inBand) BootFrom(hal.BootTarget) error {
	return errorNotImplemented
}
func (i *inBand) Firmware() (hal.FirmwareMode, error) {
	return i.common.Firmware()
}
func (i *inBand) SetFirmware(hal.FirmwareMode) error {
	return errorNotImplemented
}

// OutBand

func (o *outBand) UUID() (*uuid.UUID, error) {
	u, err := o.redfish.MachineUUID()
	if err != nil {
		return nil, err
	}
	us, err := uuid.Parse(u)
	if err != nil {
		return nil, err
	}
	return &us, nil
}
func (o *outBand) PowerState() (hal.PowerState, error) {
	return o.common.PowerState()
}
func (o *outBand) PowerOn() error {
	return errorNotImplemented
}
func (o *outBand) PowerOff() error {
	return errorNotImplemented
}

func (o *outBand) PowerReset() error {
	return errorNotImplemented
}

func (o *outBand) PowerCycle() error {
	return errorNotImplemented
}

func (o *outBand) IdentifyLEDState(hal.IdentifyLEDState) error {
	return errorNotImplemented
}

func (o *outBand) IdentifyLEDOn() error {
	return errorNotImplemented
}

func (o *outBand) IdentifyLEDOff() error {
	return errorNotImplemented
}

func (o *outBand) BootFrom(hal.BootTarget) error {
	return errorNotImplemented
}
