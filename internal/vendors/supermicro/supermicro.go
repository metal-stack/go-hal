package supermicro

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/inband"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/outband"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/pkg/api"
	goipmi "github.com/vmware/goipmi"
)

const (
	sumBin     = "sum"
	compliance = api.SMCIPMIToolCompliance
)

type (
	inBand struct {
		*inband.InBand
		sum *sum
	}
	outBand struct {
		*outband.OutBand
		sum *sum
	}
)

var (
	// errorNotImplemented for all functions that are not implemented yet
	errorNotImplemented = fmt.Errorf("not implemented yet")
)

// InBand creates an inband connection to a supermicro server.
func InBand(board *api.Board) (hal.InBand, error) {
	s, err := newSum(sumBin)
	if err != nil {
		return nil, err
	}
	ib, err := inband.New(compliance, board, true)
	if err != nil {
		return nil, err
	}
	return &inBand{
		InBand: ib,
		sum:    s,
	}, nil
}

// OutBand creates an outband connection to a supermicro server.
func OutBand(r *redfish.APIClient, board *api.Board, ip, user, password string) (hal.OutBand, error) {
	rs, err := newRemoteSum(sumBin, ip, user, password)
	if err != nil {
		return nil, err
	}
	return &outBand{
		OutBand: outband.New(r, board, compliance, ip, user, password),
		sum:     rs,
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
	return "InBand connected to Supermicro"
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
		u, err = ob.sum.uuidRemote()
		if err != nil {
			return nil, err
		}
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
	return ob.SetChassisControl(goipmi.ControlPowerDown)
}

func (ob *outBand) PowerOn() error {
	return ob.SetChassisControl(goipmi.ControlPowerUp)
}

func (ob *outBand) PowerReset() error {
	return ob.SetChassisControl(goipmi.ControlPowerHardReset)
}

func (ob *outBand) PowerCycle() error {
	return ob.SetChassisControl(goipmi.ControlPowerCycle)
}

func (ob *outBand) IdentifyLEDState(state hal.IdentifyLEDState) error {
	return ob.SetChassisIdentifyLEDState(state)
}

func (ob *outBand) IdentifyLEDOn() error {
	return ob.SetChassisIdentifyLEDOn()
}

func (ob *outBand) IdentifyLEDOff() error {
	return ob.SetChassisIdentifyLEDOff()
}

func (ob *outBand) BootFrom(bootTarget hal.BootTarget) error {
	return ob.SetBootOrder(bootTarget)
}

func (ob *outBand) Describe() string {
	return "OutBand connected to Supermicro"
}
