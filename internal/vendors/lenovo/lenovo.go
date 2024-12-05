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
	"github.com/metal-stack/go-hal/pkg/logger"
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
	bmcConnection struct {
		*inBand
	}
	bmcConnectionOutBand struct {
		*outBand
	}
)

// InBand creates an inband connection to a Lenovo server.
func InBand(board *api.Board, log logger.Logger) (hal.InBand, error) {
	ib, err := inband.New(board, true, log)
	if err != nil {
		return nil, err
	}
	return &inBand{
		InBand: ib,
	}, nil
}

// OutBand creates an outband connection to a Lenovo server.
func OutBand(r *redfish.APIClient, board *api.Board) hal.OutBand {
	return &outBand{
		OutBand: outband.ViaRedfish(r, board),
	}
}

func (ob *outBand) Close() {
	ob.Redfish.Gofish.Logout()
}

// InBand

// PowerState implements hal.InBand.
func (ib *inBand) PowerState() (hal.PowerState, error) {
	return hal.PowerOnState, nil
}

func (ib *inBand) PowerOn() error {
	return ib.IpmiTool.SetChassisControl(ipmi.ChassisControlPowerDown)
}
func (ib *inBand) PowerOff() error {
	return ib.IpmiTool.SetChassisControl(ipmi.ChassisControlPowerDown)
}
func (ib *inBand) PowerCycle() error {
	return ib.IpmiTool.SetChassisControl(ipmi.ChassisControlPowerCycle)
}

func (ib *inBand) PowerReset() error {
	return ib.IpmiTool.SetChassisControl(ipmi.ChassisControlHardReset)
}
func (o *inBand) GetIdentifyLED() (hal.IdentifyLEDState, error) {
	return hal.IdentifyLEDStateUnknown, nil
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

func (ib *inBand) BMCConnection() api.BMCConnection {
	return &bmcConnection{
		inBand: ib,
	}
}

func (c *bmcConnection) BMC() (*api.BMC, error) {
	return c.IpmiTool.BMC()
}

func (c *bmcConnection) PresentSuperUser() api.BMCUser {
	return api.BMCUser{
		Name:          "USERID",
		Id:            "2",
		ChannelNumber: 1,
	}
}

func (c *bmcConnection) SuperUser() api.BMCUser {
	return api.BMCUser{
		Name:          "root",
		Id:            "4",
		ChannelNumber: 1,
	}
}

func (c *bmcConnection) User() api.BMCUser {
	return api.BMCUser{
		Name:          "metal",
		Id:            "3",
		ChannelNumber: 1,
	}
}

func (c *bmcConnection) Present() bool {
	return c.IpmiTool.DevicePresent()
}

func (c *bmcConnection) CreateUserAndPassword(user api.BMCUser, privilege api.IpmiPrivilege) (string, error) {
	board, err := c.Board()
	if err != nil {
		return "", err
	}
	return c.IpmiTool.CreateUser(user, privilege, "", board.Vendor.PasswordConstraints(), ipmi.LowLevel)
}

func (c *bmcConnection) CreateUser(user api.BMCUser, privilege api.IpmiPrivilege, password string) error {
	_, err := c.IpmiTool.CreateUser(user, privilege, password, nil, ipmi.LowLevel)
	return err
}

func (c *bmcConnection) ChangePassword(user api.BMCUser, newPassword string) error {
	return c.IpmiTool.ChangePassword(user, newPassword, ipmi.LowLevel)
}

func (c *bmcConnection) SetUserEnabled(user api.BMCUser, enabled bool) error {
	return c.IpmiTool.SetUserEnabled(user, enabled, ipmi.LowLevel)
}

func (ib *inBand) ConfigureBIOS() (bool, error) {
	//return false, errorNotImplemented // do not throw an error to not break manual tests
	return false, nil //TODO https://github.com/metal-stack/go-hal/issues/11
}

func (ib *inBand) EnsureBootOrder(bootloaderID string) error {
	//return errorNotImplemented // do not throw an error to not break manual tests
	return nil //TODO https://github.com/metal-stack/go-hal/issues/11
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

func (o *outBand) GetIdentifyLED() (hal.IdentifyLEDState, error) {
	return o.Redfish.GetIdentifyLED()
}

func (ob *outBand) IdentifyLEDState(state hal.IdentifyLEDState) error {
	return errorNotImplemented //TODO https://github.com/metal-stack/go-hal/issues/11
}

func (ob *outBand) IdentifyLEDOn() error {
	return errorNotImplemented //TODO https://github.com/metal-stack/go-hal/issues/11
}

func (ob *outBand) IdentifyLEDOff() error {
	return errorNotImplemented //TODO https://github.com/metal-stack/go-hal/issues/11
}

func (ob *outBand) BootFrom(target hal.BootTarget) error {
	return ob.Redfish.SetBootOrder(target, vendor)
}

func (ob *outBand) Describe() string {
	return "OutBand connected to Lenovo"
}

func (ob *outBand) Console(s ssh.Session) error {
	return errorNotImplemented // https://github.com/metal-stack/go-hal/issues/11
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
	return c.Redfish.BMC()
}
