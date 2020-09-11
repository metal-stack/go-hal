package vagrant

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/console"
	"github.com/metal-stack/go-hal/internal/inband"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/outband"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/pkg/errors"
	goipmi "github.com/vmware/goipmi"
	"io"
	"os/exec"
)

var (
	// errorNotImplemented for all functions that are not implemented yet
	errorNotImplemented = fmt.Errorf("not implemented yet")
)

const (
	vendor = api.VendorVagrant
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
	ib, err := inband.New(board, false)
	if err != nil {
		return nil, err
	}
	return &inBand{
		InBand: ib,
	}, nil
}

// OutBand creates an outband connection to a vagrant VM.
func OutBand(board *api.Board, ip string, ipmiPort int, user, password string) hal.OutBand {
	return &outBand{
		OutBand: outband.ViaGoipmi(board, ip, ipmiPort, user, password),
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
	return "InBand connected to Vagrant"
}

func (ib *inBand) BMCPresentSuperUser() hal.BMCUser {
	return hal.BMCUser{}
}

func (ib *inBand) BMCSuperUser() hal.BMCUser {
	return hal.BMCUser{}
}

func (ib *inBand) BMCUser() hal.BMCUser {
	return hal.BMCUser{}
}

func (ib *inBand) BMCPresent() bool {
	return ib.IpmiTool.DevicePresent()
}

func (ib *inBand) BMCCreateUserAndPassword(user hal.BMCUser, privilege api.IpmiPrivilege, constraints api.PasswordConstraints) (string, error) {
	return ib.IpmiTool.CreateUser(user, privilege, "", &constraints, ipmi.HighLevel)
}

func (ib *inBand) BMCCreateUser(user hal.BMCUser, privilege api.IpmiPrivilege, password string) error {
	_, err := ib.IpmiTool.CreateUser(user, privilege, password, nil, ipmi.HighLevel)
	return err
}

func (ib *inBand) BMCChangePassword(user hal.BMCUser, newPassword string) error {
	return ib.IpmiTool.ChangePassword(user, newPassword, ipmi.HighLevel)
}

func (ib *inBand) BMCSetUserEnabled(user hal.BMCUser, enabled bool) error {
	return ib.IpmiTool.SetUserEnabled(user, enabled, ipmi.HighLevel)
}

func (ib *inBand) ConfigureBIOS() (bool, error) {
	return false, errorNotImplemented
}

func (ib *inBand) EnsureBootOrder(bootloaderID string) error {
	return errorNotImplemented
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
	return "OutBand connected to Vagrant"
}

func (ob *outBand) Console(s ssh.Session) error { //Virsh console
	_, err := io.WriteString(s, "Exit with '<Ctrl> 5'\n")
	if err != nil {
		return errors.Wrap(err, "failed to write to console")
	}
	ip, port, _, _ := ob.IPMIConnection()
	addr := fmt.Sprintf("%s:%d", ip, port)
	cmd := exec.Command("virsh", "console", addr, "--force")
	return console.Open(s, cmd)
}
