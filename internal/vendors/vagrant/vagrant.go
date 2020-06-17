package vagrant

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/inband"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/outband"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/pkg/api"
)

const (
	compliance = api.IPMI2Compliance
)

var (
	// errorNotImplemented for all functions that are not implemented yet
	errorNotImplemented = fmt.Errorf("not implemented yet")
)

type (
	inBand struct {
		*inband.InBand
	}
	outBand struct {
		*outband.OutBand
	}
)

// InBand creates an inband connection to a vagrant VM.
func InBand(board *api.Board) (hal.InBand, error) {
	ib, err := inband.New(compliance, board, false)
	if err != nil {
		return nil, err
	}
	return &inBand{
		InBand: ib,
	}, nil
}

// OutBand creates an outband connection to a vagrant VM.
func OutBand(r *redfish.APIClient, board *api.Board, ip, user, password string) (hal.OutBand, error) {
	ob, err := outband.New(r, board, compliance, ip, user, password)
	if err != nil {
		return nil, err
	}
	return &outBand{
		OutBand: ob,
	}, nil
}

// InBand
func (ib *inBand) PowerOff() error {
	return ib.Ipmi.SetChassisControl(ipmi.ChassisControlPowerDown)
}

func (ib *inBand) PowerCycle() error {
	return ib.Ipmi.SetChassisControl(ipmi.ChassisControlPowerCycle)
}

func (ib *inBand) PowerReset() error {
	return ib.Ipmi.SetChassisControl(ipmi.ChassisControlHardReset)
}

func (ib *inBand) IdentifyLEDState(state hal.IdentifyLEDState) error {
	return ib.Ipmi.SetChassisIdentifyLEDState(state)
}

func (ib *inBand) IdentifyLEDOn() error {
	return ib.Ipmi.SetChassisIdentifyLEDOn()
}

func (ib *inBand) IdentifyLEDOff() error {
	return ib.Ipmi.SetChassisIdentifyLEDOff()
}

func (ib *inBand) BootFrom(bootTarget hal.BootTarget) error {
	return ib.Ipmi.SetBootOrder(bootTarget)
}

func (ib *inBand) SetFirmware(hal.FirmwareMode) error {
	return errorNotImplemented //TODO
}

func (ib *inBand) Describe() string {
	return "InBand connected to Vagrant"
}

func (ib *inBand) BMCPresent() bool {
	return ib.Ipmi.DevicePresent()
}

func (ib *inBand) BMCCreateUser(username, uid string) (string, error) {
	return ib.Ipmi.CreateUser(username, uid, ipmi.AdministratorPrivilege)
}

// OutBand
func (ob *outBand) UUID() (*uuid.UUID, error) {
	u, err := ob.Redfish.MachineUUID()
	if err != nil {
		return nil, err
	}
	us, err := uuid.Parse(u)
	if err != nil {
		return nil, err
	}
	return &us, nil
}

func (ob *outBand) PowerState() (hal.PowerState, error) {
	return ob.Redfish.PowerState()
}

func (ob *outBand) PowerOff() error {
	return ob.Ipmi.SetChassisControl(ipmi.ChassisControlPowerDown)
}

func (ob *outBand) PowerOn() error {
	return ob.Ipmi.SetChassisControl(ipmi.ChassisControlPowerUp)
}

func (ob *outBand) PowerReset() error {
	return ob.Ipmi.SetChassisControl(ipmi.ChassisControlHardReset)
}

func (ob *outBand) PowerCycle() error {
	return ob.Ipmi.SetChassisControl(ipmi.ChassisControlPowerCycle)
}

func (ob *outBand) IdentifyLEDState(state hal.IdentifyLEDState) error {
	return ob.Ipmi.SetChassisIdentifyLEDState(state)
}

func (ob *outBand) IdentifyLEDOn() error {
	return ob.Ipmi.SetChassisIdentifyLEDOn()
}

func (ob *outBand) IdentifyLEDOff() error {
	return ob.Ipmi.SetChassisIdentifyLEDOff()
}

func (ob *outBand) BootFrom(bootTarget hal.BootTarget) error {
	return ob.Ipmi.SetBootOrder(bootTarget)
}

func (ob *outBand) Describe() string {
	return "OutBand connected to Vagrant"
}
