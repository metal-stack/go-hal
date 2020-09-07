package lenovo

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/inband"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/outband"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/pkg/api"
)

var (
	// errorNotImplemented for all functions that are not implemented yet
	errorNotImplemented = fmt.Errorf("not implemented yet")
)

const (
	vendor = api.VendorLenovo
)

type (
	inBand struct {
		*inband.InBand
	}
	outBand struct {
		*outband.OutBand
	}
)

// InBand creates an inband connection to a Lenovo server.
func InBand(board *api.Board) (hal.InBand, error) {
	ib, err := inband.New(board, true)
	if err != nil {
		return nil, err
	}
	return &inBand{
		InBand: ib,
	}, nil
}

// OutBand creates an outband connection to a Lenovo server.
func OutBand(r *redfish.APIClient, board *api.Board, ip string, ipmiPort int, user, password string) hal.OutBand {
	return &outBand{
		OutBand: outband.ViaRedfish(r, board, ip, ipmiPort, user, password),
	}
}

// InBand
func (ib *inBand) PowerOff() error {
	return ib.IpmiTool.SetChassisControl(ipmi.ChassisControlPowerDown)
}

func (ib *inBand) PowerCycle() error {
	return ib.IpmiTool.SetChassisControl(ipmi.ChassisControlPowerCycle)
}

func (ib *inBand) PowerReset() error {
	return ib.IpmiTool.SetChassisControl(ipmi.ChassisControlHardReset)
}

func (ib *inBand) IdentifyLEDState(state hal.IdentifyLEDState) error {
	return ib.IpmiTool.SetChassisIdentifyLEDState(state)
}

func (ib *inBand) IdentifyLEDOn() error {
	return ib.IpmiTool.SetChassisIdentifyLEDOn()
}

func (ib *inBand) IdentifyLEDOff() error {
	return ib.IpmiTool.SetChassisIdentifyLEDOff()
}

func (ib *inBand) BootFrom(bootTarget hal.BootTarget) error {
	return ib.IpmiTool.SetBootOrder(bootTarget, vendor)
}

func (ib *inBand) SetFirmware(hal.FirmwareMode) error {
	return errorNotImplemented //TODO
}

func (ib *inBand) Describe() string {
	return "InBand connected to Lenovo"
}

func (ib *inBand) BMCUser() hal.BMCUser {
	return hal.BMCUser{
		Name:          "metal",
		Uid:           "3",
		ChannelNumber: 1,
	}
}

func (ib *inBand) BMCPresent() bool {
	return ib.IpmiTool.DevicePresent()
}

func (ib *inBand) BMCCreateUser(channelNumber int, username, uid string, privilege api.IpmiPrivilege, constraints api.PasswordConstraints) (string, error) {
	return ib.IpmiTool.CreateUserRaw(channelNumber, username, uid, privilege, constraints)
}

func (ib *inBand) ConfigureBIOS() (bool, error) {
	//return false, errorNotImplemented //FIXME
	return true, nil
}

func (ib *inBand) EnsureBootOrder(bootloaderID string) error {
	//return errorNotImplemented //FIXME
	return nil
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
	return ob.Redfish.PowerOff()
}

func (ob *outBand) PowerOn() error {
	return ob.Redfish.PowerReset() // PowerOn is not supported
}

func (ob *outBand) PowerReset() error {
	return ob.Redfish.PowerReset()
}

func (ob *outBand) PowerCycle() error {
	return ob.Redfish.PowerReset() // PowerCycle is not supported
}

func (ob *outBand) IdentifyLEDState(state hal.IdentifyLEDState) error {
	//return errorNotImplemented //TODO
	return nil
}

func (ob *outBand) IdentifyLEDOn() error {
	//return errorNotImplemented //TODO
	return nil
}

func (ob *outBand) IdentifyLEDOff() error {
	//return errorNotImplemented //TODO
	return nil
}

func (ob *outBand) BootFrom(target hal.BootTarget) error {
	return ob.Redfish.SetBootOrder(target, vendor)
}

func (ob *outBand) Describe() string {
	return "OutBand connected to Lenovo"
}

func (ob *outBand) Console(s ssh.Session) error {
	return errorNotImplemented
}
