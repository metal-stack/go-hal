package supermicro

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/inband"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/outband"
	"github.com/metal-stack/go-hal/pkg/api"
	goipmi "github.com/vmware/goipmi"
)

var (
	// errorNotImplemented for all functions that are not implemented yet
	errorNotImplemented = fmt.Errorf("not implemented yet")
)

const (
	vendor = api.VendorSupermicro
	sumBin = "sum"
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

// InBand creates an inband connection to a supermicro server.
func InBand(board *api.Board) (hal.InBand, error) {
	s, err := newSum(sumBin)
	if err != nil {
		return nil, err
	}
	ib, err := inband.New(board, true)
	if err != nil {
		return nil, err
	}
	return &inBand{
		InBand: ib,
		sum:    s,
	}, nil
}

// OutBand creates an outband connection to a supermicro server.
func OutBand(board *api.Board, ip string, ipmiPort int, user, password string) (hal.OutBand, error) {
	rs, err := newRemoteSum(sumBin, ip, user, password)
	if err != nil {
		return nil, err
	}
	i, err := ipmi.New()
	if err != nil {
		return nil, err
	}
	return &outBand{
		OutBand: outband.ViaIpmi(i, board, ip, ipmiPort, user, password),
		sum:     rs,
	}, nil
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
	return "InBand connected to Supermicro"
}

func (ib *inBand) BMCUser() hal.BMCUser {
	return hal.BMCUser{
		Name:          "metal",
		Uid:           "10",
		ChannelNumber: 1,
	}
}

func (ib *inBand) BMCPresent() bool {
	return ib.IpmiTool.DevicePresent()
}

func (ib *inBand) BMCCreateUser(channelNumber int, username, uid string, privilege api.IpmiPrivilege, constraints api.PasswordConstraints) (string, error) {
	return ib.IpmiTool.CreateUser(channelNumber, username, uid, privilege, constraints)
}

func (ib *inBand) ConfigureBIOS() (bool, error) {
	return ib.sum.ConfigureBIOS()
}

func (ib *inBand) EnsureBootOrder(bootloaderID string) error {
	return ib.sum.EnsureBootOrder(bootloaderID)
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
	return ob.Goipmi(func(client *ipmi.Client) error {
		return client.Control(goipmi.ControlPowerDown)
	})
}

func (ob *outBand) PowerOn() error {
	return ob.Goipmi(func(client *ipmi.Client) error {
		return client.Control(goipmi.ControlPowerUp)
	})
}

func (ob *outBand) PowerReset() error {
	return ob.Goipmi(func(client *ipmi.Client) error {
		return client.Control(goipmi.ControlPowerHardReset)
	})
}

func (ob *outBand) PowerCycle() error {
	return ob.Goipmi(func(client *ipmi.Client) error {
		return client.Control(goipmi.ControlPowerCycle)
	})
}

func (ob *outBand) IdentifyLEDState(state hal.IdentifyLEDState) error {
	return ob.Goipmi(func(client *ipmi.Client) error {
		return client.SetChassisIdentifyLEDState(state)
	})
}

func (ob *outBand) IdentifyLEDOn() error {
	return ob.Goipmi(func(client *ipmi.Client) error {
		return client.SetChassisIdentifyLEDOn()
	})
}

func (ob *outBand) IdentifyLEDOff() error {
	return ob.Goipmi(func(client *ipmi.Client) error {
		return client.SetChassisIdentifyLEDOff()
	})
}

func (ob *outBand) BootFrom(bootTarget hal.BootTarget) error {
	return ob.Goipmi(func(client *ipmi.Client) error {
		return client.SetBootOrder(bootTarget, vendor)
	})
}

func (ob *outBand) Describe() string {
	return "OutBand connected to Supermicro"
}

func (ob *outBand) Console(s ssh.Session) error {
	return ob.IpmiTool.OpenConsole(s)
}
