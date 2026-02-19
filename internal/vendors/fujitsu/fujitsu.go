package fujitsu

import (
	"fmt"
	"strconv"

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
	vendor = api.VendorFujitsu
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

// InBand creates an inband connection to a Fujitsu server.
func InBand(board *api.Board, log logger.Logger) (hal.InBand, error) {
	ib, err := inband.New(board, true, log)
	if err != nil {
		return nil, err
	}
	return &inBand{
		InBand: ib,
	}, nil
}

// OutBand creates an outband connection to a Fujitsu server.
func OutBand(r *redfish.APIClient, board *api.Board) hal.OutBand {
	r.SetETagRequired(true)
	return &outBand{
		OutBand: outband.ViaRedfish(r, board),
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
	return errorNotImplemented
}

func (ib *inBand) Describe() string {
	return "InBand connected to Fujitsu"
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
		ChannelNumber: 2,
	}
}

func (c *bmcConnection) SuperUser() api.BMCUser {
	return api.BMCUser{
		Name:          "root",
		Id:            "4",
		ChannelNumber: 2,
	}
}

func (c *bmcConnection) User() api.BMCUser {
	return api.BMCUser{
		Name:          "metal",
		Id:            "3",
		ChannelNumber: 2,
	}
}

func (c *bmcConnection) Present() bool {
	return c.IpmiTool.DevicePresent()
}

func (c *bmcConnection) CreateUserAndPassword(user api.BMCUser, privilege api.IpmiPrivilege) (string, error) {
	password_constraints := c.Board().Vendor.PasswordConstraints()
	password_constraints.Length = 12
	s, err := c.IpmiTool.CreateUser(user, privilege, "", password_constraints, ipmi.HighLevel)
	if err != nil {
		return "", err
	}

	err_perm := c.syncRedfishPermissions(user, privilege)
	if err_perm != nil {
		return "", err_perm
	}
	return s, nil
}

func (c *bmcConnection) CreateUser(user api.BMCUser, privilege api.IpmiPrivilege, password string) error {
	_, err := c.IpmiTool.CreateUser(user, privilege, password, nil, ipmi.HighLevel)
	if err != nil {
		return err
	}
	return c.syncRedfishPermissions(user, privilege)
}

func (c *bmcConnection) NeedsPasswordChange(user api.BMCUser, password string) (bool, error) {
	return c.IpmiTool.NeedsPasswordChange(user, password)
}

func (c *bmcConnection) ChangePassword(user api.BMCUser, newPassword string) error {
	return c.IpmiTool.ChangePassword(user, newPassword, ipmi.HighLevel)
}

func (c *bmcConnection) SetUserEnabled(user api.BMCUser, enabled bool) error {
	err := c.IpmiTool.SetUserEnabled(user, enabled, ipmi.HighLevel)
	if err != nil {
		return err
	}
	return c.syncRedfishPermissions(user, api.UserPrivilege)
}

func (c *bmcConnection) syncRedfishPermissions(user api.BMCUser, privilege api.IpmiPrivilege) error {
	// 1. Parse the User ID
	if user.Id == "" {
		return fmt.Errorf("user ID is empty, cannot set Redfish permissions")
	}

	userIDInt, err := strconv.Atoi(user.Id)
	if err != nil {
		return fmt.Errorf("failed to parse user ID %q: %w", user.Id, err)
	}

	// IPMI raw userID = <Redfish UserID> - 1
	targetUserID := userIDInt - 1
	userIDHex := fmt.Sprintf("0x%02x", targetUserID)

	// 2. Map standard IPMI Privileges to Fujitsu OEM Redfish Role values
	var roleHex string
	enabled := "0x01" // Enabled
	switch privilege {
	case api.AdministratorPrivilege:
		roleHex = "0x02" // Administrator
	case api.OperatorPrivilege:
		roleHex = "0x01" // Operator
	case api.UserPrivilege:
		roleHex = "0x03" // Read-Only
	default:
		roleHex = "0x00" // No Access
		enabled = "0x00" // Disabled
	}

	// 3. Set Role (0x81 0x1D feature code)
	// ipmitool raw 0x2e 0xe0 0x80 0x28 0x00 0x02 <userID> 0x81 0x1D 0x01 <RoleValue>
	// Example for user ID 3 (which becomes 2 in hex) and Administrator role (0x02):
	// ipmitool raw 0x2e 0xe0 0x80 0x28 0x00 0x02 0x02 0x81 0x1D 0x01 0x02
	setRoleCmd := []string{
		"raw", "0x2e", "0xe0", "0x80", "0x28", "0x00",
		"0x02", userIDHex, "0x81", "0x1D", "0x01", roleHex,
	}

	_, err = c.IpmiTool.Run(setRoleCmd...)
	if err != nil {
		return fmt.Errorf("failed to set fujitsu redfish role for user %d: %w", userIDInt, err)
	}

	// 4. Enable Access (0x81 0x1D feature code)
	// ipmitool raw 0x2e 0xe0 0x80 0x28 0x00 0x02 <userID> 0x81 0x1D 0x01 <0x01 for enable, 0x00 for disable>
	// Example to enable access for user ID 3:
	// ipmitool raw 0x2e 0xe0 0x80 0x28 0x00 0x02 0x02 0x81 0x1D 0x01 0x01
	enableAccessCmd := []string{
		"raw", "0x2e", "0xe0", "0x80", "0x28", "0x00",
		"0x02", userIDHex, "0x80", "0x1D", "0x01", enabled,
	}

	_, err = c.IpmiTool.Run(enableAccessCmd...)
	if err != nil {
		return fmt.Errorf("failed to enable fujitsu redfish access for user %d: %w", userIDInt, err)
	}

	return nil
}

func (ib *inBand) ConfigureBIOS() (bool, error) {
	// return errorNotImplemented
	//return false, errorNotImplemented // do not throw an error to not break manual tests
	return false, nil //TODO https://github.com/metal-stack/go-hal/issues/11
}

func (ib *inBand) EnsureBootOrder(bootloaderID string) error {
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
	return ob.Redfish.PowerOn()
}

func (ob *outBand) PowerReset() error {
	return ob.Redfish.PowerReset()
}

func (ob *outBand) PowerCycle() error {
	system, err := ob.Redfish.GetSystem()
	if err != nil {
		return err
	}

	// Construct OEM action path
	oemPath := fmt.Sprintf("%s/Actions/Oem/FTSComputerSystem.Reset", system.ODataID)

	body := map[string]interface{}{
		"FTSResetType": "PowerCycle",
	}

	err = system.Post(oemPath, body)
	return err
}

func (ob *outBand) IdentifyLEDState(state hal.IdentifyLEDState) error {
	return ob.Redfish.SetChassisIdentifyLEDState(state)
}

func (ob *outBand) IdentifyLEDOn() error {
	return ob.Redfish.SetChassisIdentifyLEDOn()
}

func (ob *outBand) IdentifyLEDOff() error {
	return ob.Redfish.SetChassisIdentifyLEDOff()
}

func (ob *outBand) BootFrom(target hal.BootTarget) error {
	// On Fujitsu for BootSourceOverrideTarget = "BiosSetup" BootSourceOverrideEnabled is restricted to "Once"
	return ob.Redfish.SetBootOrder(target)
}

func (ob *outBand) Describe() string {
	return "OutBand connected to Fujitsu"
}

func (ob *outBand) Console(s ssh.Session) error {
	return errorNotImplemented // TODO use the same implementation as for dell after it's merged
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
