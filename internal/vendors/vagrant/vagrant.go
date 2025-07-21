package vagrant

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/console"
	"github.com/metal-stack/go-hal/internal/inband"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/outband"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/metal-stack/go-hal/pkg/logger"
	goipmi "github.com/vmware/goipmi"
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
	bmcConnection struct {
		*inBand
	}
	bmcConnectionOutBand struct {
		*outBand
	}
)

// InBand creates an inband connection to a vagrant VM.
func InBand(board *api.Board, log logger.Logger) (hal.InBand, error) {
	ib, err := inband.New(board, false, log)
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
	return nil
}

func (ib *inBand) IdentifyLEDOn() error {
	return nil
}

func (ib *inBand) IdentifyLEDOff() error {
	return nil
}

func (ib *inBand) BootFrom(bootTarget hal.BootTarget) error {
	return ib.IpmiTool.SetBootOrder(bootTarget, vendor)
}

func (ib *inBand) SetFirmware(hal.FirmwareMode) error {
	return nil
}

func (ib *inBand) Describe() string {
	return "InBand connected to Vagrant"
}

func (ib *inBand) BMCConnection() api.BMCConnection {
	return &bmcConnection{
		inBand: ib,
	}
}

func (c *bmcConnection) BMC() (*api.BMC, error) {
	return api.VagrantBoard.BMC, nil
}

func (c *bmcConnection) PresentSuperUser() api.BMCUser {
	return api.BMCUser{}
}

func (c *bmcConnection) SuperUser() api.BMCUser {
	return api.BMCUser{}
}

func (c *bmcConnection) User() api.BMCUser {
	return api.BMCUser{}
}

func (c *bmcConnection) Present() bool {
	return false
}

func (c *bmcConnection) CreateUserAndPassword(user api.BMCUser, privilege api.IpmiPrivilege) (string, error) {
	return "", nil
}

func (c *bmcConnection) CreateUser(user api.BMCUser, privilege api.IpmiPrivilege, password string) error {
	return nil
}

func (c *bmcConnection) TestUserPassword(user api.BMCUser, passwordSize int, password string) error {
	return nil
}

func (c *bmcConnection) ChangePassword(user api.BMCUser, newPassword string) error {
	return nil
}

func (c *bmcConnection) SetUserEnabled(user api.BMCUser, enabled bool) error {
	return nil
}

func (ib *inBand) ConfigureBIOS() (bool, error) {
	return false, nil
}

func (ib *inBand) EnsureBootOrder(bootloaderID string) error {
	return nil
}

// OutBand
func (ob *outBand) UUID() (*uuid.UUID, error) {
	return nil, nil
}

func (ob *outBand) PowerState() (hal.PowerState, error) {
	return hal.PowerOnState, nil
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
	return nil
}

func (ob *outBand) IdentifyLEDOn() error {
	return nil
}

func (ob *outBand) IdentifyLEDOff() error {
	return nil
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
		return fmt.Errorf("failed to write to console %w", err)
	}
	ip, port, _, _ := ob.IPMIConnection()
	addr := fmt.Sprintf("%s:%d", ip, port)
	cmd := exec.Command("virsh", "console", addr, "--force")
	return console.Open(s, cmd)
}

func (ob *outBand) UpdateBIOS(url string) error {
	return nil
}

func (ob *outBand) UpdateBMC(url string) error {
	return nil
}

func (ob *outBand) BMCConnection() api.OutBandBMCConnection {
	return &bmcConnectionOutBand{
		outBand: ob,
	}
}

func (c *bmcConnectionOutBand) BMC() (*api.BMC, error) {
	return api.VagrantBoard.BMC, nil
}
