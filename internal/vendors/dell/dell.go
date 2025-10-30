package dell

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"

	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/console"
	"github.com/metal-stack/go-hal/internal/inband"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/outband"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/metal-stack/go-hal/pkg/logger"

	gofish "github.com/stmcginnis/gofish/redfish"
)

var (
	// errorNotImplemented for all functions that are not implemented yet
	errorNotImplemented = fmt.Errorf("not implemented yet")
)

const (
	vendor  = api.VendorDell
	sshPort = 22 // default SSH port for Dell BMCs
)

type (
	inBand struct {
		*inband.InBand
	}
	outBand struct {
		*outband.OutBand
		log logger.Logger
	}
	bmcConnection struct {
		*inBand
	}
	bmcConnectionOutBand struct {
		*outBand
	}
)

// InBand creates an inband connection to a Dell server.
func InBand(board *api.Board, log logger.Logger) (hal.InBand, error) {
	ib, err := inband.New(board, false, log)
	if err != nil {
		return nil, err
	}
	return &inBand{
		InBand: ib,
	}, nil
}

// OutBand creates an outband connection to a Dell server.
func OutBand(r *redfish.APIClient, board *api.Board, user, password, ip string, log logger.Logger) hal.OutBand {
	return &outBand{
		OutBand: outband.ViaRedfishPlusSSH(r, board, user, password, ip, sshPort),
		log:     log,
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
	return "InBand connected to Dell"
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
		Name:          "superuser",
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
	return c.IpmiTool.CreateUser(user, privilege, "", c.Board().Vendor.PasswordConstraints(), ipmi.HighLevel)
}

func (c *bmcConnection) CreateUser(user api.BMCUser, privilege api.IpmiPrivilege, password string) error {
	_, err := c.IpmiTool.CreateUser(user, privilege, password, nil, ipmi.HighLevel)
	return err
}

func (c *bmcConnection) NeedsPasswordChange(user api.BMCUser, password string) (bool, error) {
	return c.IpmiTool.NeedsPasswordChange(user, password)
}

func (c *bmcConnection) ChangePassword(user api.BMCUser, newPassword string) error {
	return c.IpmiTool.ChangePassword(user, newPassword, ipmi.HighLevel)
}

func (c *bmcConnection) SetUserEnabled(user api.BMCUser, enabled bool) error {
	return c.IpmiTool.SetUserEnabled(user, enabled, ipmi.HighLevel)
}

func (ib *inBand) ConfigureBIOS() (bool, error) {
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
	return ob.Redfish.PowerReset() // PowerOn is not supported
}

func (ob *outBand) PowerReset() error {
	return ob.Redfish.PowerReset()
}

func (ob *outBand) PowerCycle() error {
	return ob.Redfish.PowerReset() // PowerCycle is not supported
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

func cutVersion(ver string) string {
	parts := strings.Split(ver, ".")
	if len(parts) > 3 {
		return strings.Join(parts[:3], ".")
	}
	return ver
}

func (ob *outBand) BootFrom(target hal.BootTarget) error {
	// cut the version to ensure it adheres to semver
	currentBiosVersion, errVersion := semver.NewVersion(cutVersion(ob.Board().BiosVersion))
	if errVersion != nil {
		ob.log.Infow("failed to parse BIOS version '%s': %v, falling back to legacy boot device setup", ob.Board().BiosVersion, errVersion)
	}
	// Dell fixed a bug in 2.75.75.75 BIOS version, we can use the normal way now
	// As we drop the last part of the version we need to check for strictly greater than 2.75.75
	neededBiosVersionCheck, errConstraint := semver.NewConstraint("> 2.75.75")
	if errConstraint != nil {
		ob.log.Infow("failed to parse BIOS version constraint: %w, falling back to legacy boot device setup", errConstraint)
	}
	if errVersion == nil && errConstraint == nil && neededBiosVersionCheck.Check(currentBiosVersion) {
		return ob.Redfish.SetBootTarget(target)
	}
	// Dell has a bug in the implementation of setting the BootSourceOverrideEnabled to "Continuous". It only survives a single reboot
	// Mentioned here under 159467: https://www.dell.com/support/manuals/en-us/dell-dss-7500/idrac8_2.75.75.75_rn/automation-api-and-cli?guid=guid-156e2423-80df-46f9-9d00-cb1018c6e227&lang=en-us
	// And in this changelog: https://www.dell.com/support/home/en-us/drivers/driversdetails?driverid=krcxx

	// As a workaround we modify the BootOrder directly
	bootOptions, err := ob.Redfish.GetBootOptions()
	if err != nil {
		return err
	}

	switch target {
	case hal.BootTargetBIOS:
		return ob.Redfish.SetBootTarget(target)
	case hal.BootTargetDisk:
		var hdOptions []*gofish.BootOption
		for _, option := range bootOptions {
			if strings.HasPrefix(option.UefiDevicePath, "HD(") {
				hdOptions = append(hdOptions, option)
			}
		}
		if len(hdOptions) == 0 {
			return fmt.Errorf("no hard disk boot option found")
		}
		return ob.Redfish.SetBootOrder(hdOptions)
	case hal.BootTargetPXE:
		fallthrough
	default:
		var nicOptions []*gofish.BootOption
		for _, option := range bootOptions {
			if strings.Contains(option.DisplayName, "NIC") || strings.Contains(option.DisplayName, "PXE") {
				nicOptions = append(nicOptions, option)
			}
		}
		if len(nicOptions) == 0 {
			return fmt.Errorf("no PXE boot option found")
		}
		return ob.Redfish.SetBootOrder(nicOptions)
	}
}

func (ob *outBand) Describe() string {
	return "OutBand connected to Dell"
}

func (ob *outBand) Console(s ssh.Session) error {
	return console.OverSSH(s, ob.GetUsername(), ob.GetPassword(), ob.GetIP(), ob.GetSSHPort(), ob.log)
}

func (ob *outBand) UpdateBIOS(url string) error {
	return ob.Redfish.UpdateFirmware(url)
}

func (ob *outBand) UpdateBMC(url string) error {
	return ob.Redfish.UpdateFirmware(url)
}

func (ob *outBand) BMCConnection() api.OutBandBMCConnection {
	return &bmcConnectionOutBand{
		outBand: ob,
	}
}

func (c *bmcConnectionOutBand) BMC() (*api.BMC, error) {
	return c.Redfish.BMC()
}
