package generic

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/inband"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/metal-stack/go-hal/pkg/logger"
)

var (
	// errorNotImplemented for all functions that are not implemented yet
	errorNotImplemented = fmt.Errorf("not implemented yet")
)

const (
	defaultChannel = 1
)

type (
	inBand struct {
		*inband.InBand
	}
	bmcConnection struct {
		*inBand
	}
)

// InBand creates an inband connection to a server using a generic,
// vendor-agnostic IPMI implementation that does not rely on any
// vendor-specific commands.
func InBand(board *api.Board, log logger.Logger) (hal.InBand, error) {
	ib, err := inband.New(board, true, log)
	if err != nil {
		return nil, err
	}
	return &inBand{
		InBand: ib,
	}, nil
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
	return "InBand connected via generic IPMI"
}

func (ib *inBand) BMCConnection() api.BMCConnection {
	return &bmcConnection{
		inBand: ib,
	}
}

func (c *bmcConnection) BMC() (*api.BMC, error) {
	return c.IpmiTool.BMC()
}

// users returns the users currently configured on the BMC by parsing the
// output of "ipmitool user list <channel>".
func (c *bmcConnection) users() []api.BMCUser {
	output, err := c.IpmiTool.Run("user", "list", strconv.Itoa(defaultChannel))
	if err != nil {
		return nil
	}
	return parseUsers(output)
}

// PresentSuperUser returns the details of the currently present BMC admin
// user, detected at runtime. It falls back to a common default if the user
// list cannot be determined.
func (c *bmcConnection) PresentSuperUser() api.BMCUser {
	for _, u := range c.users() {
		if strings.Contains(strings.ToUpper(u.Name), "ADMIN") {
			return u
		}
	}
	return api.BMCUser{
		Name:          "ADMIN",
		Id:            "1",
		ChannelNumber: defaultChannel,
	}
}

// SuperUser returns the details of the preset metal bmc superuser.
func (c *bmcConnection) SuperUser() api.BMCUser {
	return api.BMCUser{
		Name:          "root",
		Id:            "4",
		ChannelNumber: defaultChannel,
	}
}

// User returns the details of the preset metal bmc user. A free user id is
// chosen at runtime so that an already existing user is not overwritten.
func (c *bmcConnection) User() api.BMCUser {
	const defaultID = "3"
	used := make(map[string]struct{})
	for _, u := range c.users() {
		used[u.Id] = struct{}{}
	}
	id := defaultID
	for i := 2; i <= 16; i++ {
		candidate := strconv.Itoa(i)
		if _, taken := used[candidate]; !taken {
			id = candidate
			break
		}
	}
	return api.BMCUser{
		Name:          "metal",
		Id:            id,
		ChannelNumber: defaultChannel,
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
	// The IPMI 2.0 standard does not expose any mechanism to configure the
	// BIOS (e.g. enabling UEFI or disabling CSM). This is a firmware-specific
	// operation that requires vendor tooling. It is reported as a no-op so
	// that provisioning does not break on otherwise standard-compliant BMCs.
	return false, nil
}

func (ib *inBand) EnsureBootOrder(bootloaderID string) error {
	// The bootloaderID references a single UEFI boot entry (a UEFI variable).
	// Reordering UEFI boot entries is done via UEFI runtime variables, which
	// are not part of the IPMI 2.0 standard: IPMI's only boot control is the
	// "Set System Boot Options" command (see BootFrom/SetBootOrder), which
	// selects a boot device class (PXE/HDD/CD), not an individual UEFI entry.
	//
	// Since this cannot be expressed through a standard-compliant BMC, it is
	// reported as a no-op rather than an error (an error here is fatal for
	// provisioning in metal-hammer). Vendors that need real boot-order
	// control must implement it via their own tooling (e.g. Redfish or a
	// vendor BIOS tool).
	return nil
}

// parseUsers parses the output of "ipmitool user list <channel>" which is a
// fixed-width table of the form:
//
//	ID  Name       Callin  Link Auth  IPMI Msg  Channel Priv Limit
//	1   ADMIN      false   false      true       ADMINISTRATOR
//	2   USERID     false   false      true       ADMINISTRATOR
func parseUsers(output string) []api.BMCUser {
	var users []api.BMCUser
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		// A data row starts with a numeric user id and has at least the id
		// and the username. Skip the header and empty lines.
		if len(fields) < 2 {
			continue
		}
		if _, err := strconv.Atoi(fields[0]); err != nil {
			continue
		}
		users = append(users, api.BMCUser{
			Name:          fields[1],
			Id:            fields[0],
			ChannelNumber: defaultChannel,
		})
	}
	return users
}
