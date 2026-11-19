package generic

import (
	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/outband"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/metal-stack/go-hal/pkg/logger"
	goipmi "github.com/vmware/goipmi"
)

const (
	vendor = api.VendorUnknown
)

type (
	outBand struct {
		*outband.OutBand
	}
	bmcConnectionOutBand struct {
		*outBand
	}
)

// OutBand creates an outband connection to a server using a generic,
// vendor-agnostic IPMI implementation that does not use any Redfish commands.
func OutBand(board *api.Board, ip string, ipmiPort int, user, password string, log logger.Logger) (hal.OutBand, error) {
	i, err := ipmi.NewOutBand(ip, ipmiPort, user, password, log)
	if err != nil {
		return nil, err
	}
	ob := outband.ViaGoipmi(board, ip, ipmiPort, user, password)
	ob.IpmiTool = i
	return &outBand{
		OutBand: ob,
	}, nil
}

func (ob *outBand) UUID() (*uuid.UUID, error) {
	u, err := ob.IpmiTool.MachineUUID()
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
	var state hal.PowerState
	err := ob.Goipmi(func(client *ipmi.Client) error {
		var err error
		state, err = client.PowerState()
		return err
	})
	return state, err
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
	return "OutBand connected via generic IPMI"
}

func (ob *outBand) Console(s ssh.Session) error {
	return ob.IpmiTool.OpenConsole(s)
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
	return c.IpmiTool.BMC()
}
