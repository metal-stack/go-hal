package ipmi

// This package implements the small subset of IPMI that go-hal needs on top of
// github.com/bougou/go-ipmi. It replaces the former shelling out to `ipmitool`
// (both in-band and out-of-band).

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	goclient "github.com/bougou/go-ipmi/pkg/client"
	"github.com/bougou/go-ipmi/pkg/command/app"
	"github.com/bougou/go-ipmi/pkg/command/chassis"
	"github.com/bougou/go-ipmi/pkg/types"
	"github.com/gliderlabs/ssh"
	"github.com/sethvargo/go-password/password"

	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/metal-stack/go-hal/pkg/logger"
)

// defaultContext is used for all IPMI commands. The bougou client applies its
// own timeouts and retries per interface.
var defaultContext = context.Background()

// IpmiTool defines methods to interact with IPMI
type IpmiTool interface {
	DevicePresent() bool
	CreateUser(user api.BMCUser, privilege api.IpmiPrivilege, password string, constraints *api.PasswordConstraints) (pwd string, err error)
	ChangePassword(user api.BMCUser, newPassword string) error
	NeedsPasswordChange(user api.BMCUser, password string) (b bool, e error)
	SetUserEnabled(user api.BMCUser, enabled bool) error
	GetLanConfig() (LanConfig, error)
	SetBootOrder(target hal.BootTarget, vendor api.Vendor) error
	SetChassisControl(ChassisControlFunction) error
	SetChassisIdentifyLEDState(hal.IdentifyLEDState) error
	SetChassisIdentifyLEDOn() error
	SetChassisIdentifyLEDOff() error
	GetFru() (Fru, error)
	GetSession() (Session, error)
	BMC() (*api.BMC, error)
	OpenConsole(s ssh.Session) error
}

// Client implements IpmiTool on top of the native go-ipmi client. Depending on
// how it is created it talks to a local BMC (/dev/ipmi*) or a remote BMC via
// lanplus.
type Client struct {
	client *goclient.Client
	log    logger.Logger

	connectOnce sync.Once
	connectErr  error
}

var _ IpmiTool = (*Client)(nil)

// LanConfig contains the config of IPMI.
type LanConfig struct {
	IP  string
	Mac string
}

func (l *LanConfig) String() string {
	return fmt.Sprintf("ip: %s mac:%s", l.IP, l.Mac)
}

// Session holds information about the current IPMI session
type Session struct {
	UserID    string
	Privilege string
}

// Fru contains Field Replaceable Unit information
type Fru struct {
	ChassisPartNumber   string
	ChassisPartSerial   string
	BoardMfg            string
	BoardMfgSerial      string
	BoardPartNumber     string
	ProductManufacturer string
	ProductPartNumber   string
	ProductSerial       string
}

// New creates a new in-band IpmiTool talking to the local BMC via /dev/ipmi*.
func New(log logger.Logger) (IpmiTool, error) {
	c, err := goclient.NewOpenClient()
	if err != nil {
		return nil, fmt.Errorf("unable to create in-band ipmi client: %w", err)
	}
	return &Client{
		client: c,
		log:    log,
	}, nil
}

// NewOutBand creates a new IpmiTool talking to a remote BMC via lanplus.
func NewOutBand(ip string, port int, user, password string, log logger.Logger) (IpmiTool, error) {
	c, err := newLanplusClient(ip, port, user, password)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: c,
		log:    log,
	}, nil
}

// OpenClientConnection opens a new lanplus connection to a remote BMC.
func OpenClientConnection(ip string, port int, user, password string) (*Client, error) {
	c, err := newLanplusClient(ip, port, user, password)
	if err != nil {
		return nil, err
	}
	client := &Client{
		client: c,
		log:    logger.New(),
	}
	if err := client.connect(); err != nil {
		return nil, err
	}
	return client, nil
}

func newLanplusClient(ip string, port int, user, password string) (*goclient.Client, error) {
	c, err := goclient.NewClient(ip, port, user, password)
	if err != nil {
		return nil, fmt.Errorf("unable to create ipmi client for %s:%d: %w", ip, port, err)
	}
	c.WithInterface(goclient.InterfaceLanplus)
	return c, nil
}

// connect establishes the IPMI session on first use.
func (c *Client) connect() error {
	c.connectOnce.Do(func() {
		c.connectErr = c.client.Connect(defaultContext)
	})
	return c.connectErr
}

// Close closes the IPMI session.
func (c *Client) Close() error {
	return c.client.Close(defaultContext)
}

// DevicePresent returns true if the IPMI device is present, which is required to talk to the BMC
func (c *Client) DevicePresent() bool {
	const ipmiDevicePrefix = "/dev/ipmi*"
	matches, err := filepath.Glob(ipmiDevicePrefix)
	if err != nil {
		return false
	}
	return len(matches) > 0
}

// Control executes the given chassis control function.
func (c *Client) Control(fn ChassisControlFunction) error {
	return c.SetChassisControl(fn)
}

// SetChassisControl executes the given chassis control function
func (c *Client) SetChassisControl(fn ChassisControlFunction) error {
	if err := c.connect(); err != nil {
		return err
	}
	if _, err := c.client.ChassisControl(defaultContext, chassis.ChassisControl(fn)); err != nil {
		return fmt.Errorf("unable to set chassis control function:%X %w", fn, err)
	}
	return nil
}

// SetChassisIdentifyLEDState sets the chassis identify LED to given state
func (c *Client) SetChassisIdentifyLEDState(state hal.IdentifyLEDState) error {
	switch state {
	case hal.IdentifyLEDStateOn:
		return c.SetChassisIdentifyLEDOn()
	case hal.IdentifyLEDStateOff:
		return c.SetChassisIdentifyLEDOff()
	case hal.IdentifyLEDStateUnknown:
		fallthrough
	default:
		return fmt.Errorf("unknown identify LED state: %s", state)
	}
}

// SetChassisIdentifyLEDOn turns on the chassis identify LED
func (c *Client) SetChassisIdentifyLEDOn() error {
	return c.setChassisIdentifyLED(true)
}

// SetChassisIdentifyLEDOff turns off the chassis identify LED
func (c *Client) SetChassisIdentifyLEDOff() error {
	return c.setChassisIdentifyLED(false)
}

func (c *Client) setChassisIdentifyLED(forceOn bool) error {
	if err := c.connect(); err != nil {
		return err
	}
	if _, err := c.client.ChassisIdentify(defaultContext, 0, forceOn); err != nil {
		return fmt.Errorf("unable to set chassis identify LED: %w", err)
	}
	return nil
}

// SetBootOrder persistently sets the boot order to given target
func (c *Client) SetBootOrder(target hal.BootTarget, vendor api.Vendor) error {
	if err := c.connect(); err != nil {
		return err
	}
	if err := c.client.SetBootParamBootFlags(defaultContext, BootFlags(target, vendor)); err != nil {
		return fmt.Errorf("unable to persistently set boot order:%s %w", target, err)
	}
	return nil
}

// BootFlags returns the boot flags needed to set the given boot order according
// to the given vendor.
func BootFlags(target hal.BootTarget, vendor api.Vendor) *types.BootOptionParam_BootFlags {
	flags := &types.BootOptionParam_BootFlags{
		BootFlagsValid: true,
		BIOSBootType:   types.BIOSBootTypeEFI,
	}

	switch target {
	case hal.BootTargetPXE:
		flags.Persist = true
		flags.BootDeviceSelector = types.BootDeviceSelectorForcePXE
	case hal.BootTargetDisk:
		flags.Persist = true
		switch vendor {
		case api.VendorSupermicro, api.VendorNovarion:
			// Supermicro/Novarion use a different selector for the hard disk
			flags.BootDeviceSelector = types.BootDeviceSelector(0x09)
		default:
			flags.BootDeviceSelector = types.BootDeviceSelectorForceHardDrive
		}
	case hal.BootTargetBIOS:
		// booting into the BIOS setup applies to the next boot only
		flags.BootDeviceSelector = types.BootDeviceSelectorForceBIOSSetup
	}

	return flags
}

// GetLanConfig returns the LAN config
func (c *Client) GetLanConfig() (LanConfig, error) {
	if err := c.connect(); err != nil {
		return LanConfig{}, err
	}
	lan, err := c.client.GetLanConfig(defaultContext, 1)
	if err != nil {
		return LanConfig{}, fmt.Errorf("unable to get lan config: %w", err)
	}
	return LanConfig{
		IP:  lan.IP.String(),
		Mac: lan.MAC.String(),
	}, nil
}

// GetFru returns the Field Replaceable Unit information
func (c *Client) GetFru() (Fru, error) {
	if err := c.connect(); err != nil {
		return Fru{}, err
	}
	fru, err := c.client.GetFRU(defaultContext, 0x00, "Builtin FRU")
	if err != nil {
		return Fru{}, fmt.Errorf("unable to get fru: %w", err)
	}

	f := Fru{}
	if a := fru.ChassisInfoArea; a != nil {
		f.ChassisPartNumber = types.FRUFieldString(a.PartNumberTypeLength, a.PartNumber)
		f.ChassisPartSerial = types.FRUFieldString(a.SerialNumberTypeLength, a.SerialNumber)
	}
	if a := fru.BoardInfoArea; a != nil {
		f.BoardMfg = types.FRUFieldString(a.ManufacturerTypeLength, a.Manufacturer)
		f.BoardMfgSerial = types.FRUFieldString(a.SerialNumberTypeLength, a.SerialNumber)
		f.BoardPartNumber = types.FRUFieldString(a.PartNumberTypeLength, a.PartNumber)
	}
	if a := fru.ProductInfoArea; a != nil {
		f.ProductManufacturer = types.FRUFieldString(a.ManufacturerTypeLength, a.Manufacturer)
		f.ProductPartNumber = types.FRUFieldString(a.PartModelTypeLength, a.PartModel)
		f.ProductSerial = types.FRUFieldString(a.SerialNumberTypeLength, a.SerialNumber)
	}
	return f, nil
}

// GetSession returns the session
func (c *Client) GetSession() (Session, error) {
	if err := c.connect(); err != nil {
		return Session{}, err
	}
	res, err := c.client.GetCurrentSessionInfo(defaultContext)
	if err != nil {
		return Session{}, fmt.Errorf("unable to get session info: %w", err)
	}
	return Session{
		UserID:    strconv.Itoa(int(res.UserID)),
		Privilege: res.OperatingPrivilegeLevel.String(),
	}, nil
}

// BMC returns the BMC struct
func (c *Client) BMC() (*api.BMC, error) {
	lan, err := c.GetLanConfig()
	if err != nil {
		return nil, err
	}
	fru, err := c.GetFru()
	if err != nil {
		// FIXME
		c.log.Errorw("unable to get fru:%s", err)
		// return nil, err
	}
	firmwareRevision, err := c.firmwareRevision()
	if err != nil {
		// FIXME
		c.log.Errorw("unable to get bmcinfo:%s", err)
		// return nil, err
	}
	return &api.BMC{
		IP:                  lan.IP,
		MAC:                 lan.Mac,
		BoardMfg:            fru.BoardMfg,
		BoardMfgSerial:      fru.BoardMfgSerial,
		BoardPartNumber:     fru.BoardPartNumber,
		ChassisPartNumber:   fru.ChassisPartNumber,
		ChassisPartSerial:   fru.ChassisPartSerial,
		ProductManufacturer: fru.ProductManufacturer,
		ProductPartNumber:   fru.ProductPartNumber,
		ProductSerial:       fru.ProductSerial,
		FirmwareRevision:    firmwareRevision,
	}, nil
}

func (c *Client) firmwareRevision() (string, error) {
	if err := c.connect(); err != nil {
		return "", err
	}
	res, err := c.client.GetDeviceID(defaultContext)
	if err != nil {
		return "", fmt.Errorf("unable to get device id: %w", err)
	}
	return fmt.Sprintf("%d.%02d", res.MajorFirmwareRevision, res.MinorFirmwareRevision), nil
}

// CreateUser creates an IPMI user with given privilege level and either the given password or - if empty - a generated one with respect to the given password constraints
func (c *Client) CreateUser(user api.BMCUser, privilege api.IpmiPrivilege, password string, pc *api.PasswordConstraints) (string, error) {
	uid, cn, err := userIDs(user)
	if err != nil {
		return "", err
	}
	if err := c.connect(); err != nil {
		return "", err
	}

	if _, err := c.client.SetUsername(defaultContext, uid, user.Name); err != nil {
		return "", fmt.Errorf("failed set username for user %s with id %s: %w", user.Name, user.Id, err)
	}

	pw, err := c.changePassword(user, password, pc)
	if err != nil {
		return "", err
	}

	req := &app.SetUserAccessRequest{
		EnableChanging:      true,
		EnableLinkAuth:      true,
		EnableIPMIMessaging: true,
		ChannelNumber:       cn,
		UserID:              uid,
		MaxPrivLevel:        uint8(privilege),
	}
	if _, err := c.client.SetUserAccess(defaultContext, req); err != nil {
		return "", fmt.Errorf("failed to set privilege %d for user %s with id %s: %w", privilege, user.Name, user.Id, err)
	}

	// Enable SOL payload access for the user. The go-ipmi wrapper method does not
	// pass the request through, so issue the exchange directly.
	payloadAccess := &app.SetUserPayloadAccessRequest{
		ChannelNumber:  cn,
		UserID:         uid,
		Operation:      app.SetUserPayloadAccessOperationEnable,
		PayloadTypeSOL: true,
	}
	if err := c.client.Exchange(defaultContext, payloadAccess, &app.SetUserPayloadAccessResponse{}); err != nil {
		return "", fmt.Errorf("failed to set enable user SOL payload access for user %s with id %s: %w", user.Name, user.Id, err)
	}

	return pw, nil
}

// ChangePassword of the given user
func (c *Client) ChangePassword(user api.BMCUser, newPassword string) error {
	_, err := c.changePassword(user, newPassword, nil)
	return err
}

// SetUserEnabled enable the given user
func (c *Client) SetUserEnabled(user api.BMCUser, enabled bool) error {
	if err := c.connect(); err != nil {
		return err
	}
	return c.setUserEnabled(user, enabled)
}

// NeedsPasswordChange checks if the given password matches the stored one.
func (c *Client) NeedsPasswordChange(user api.BMCUser, password string) (bool, error) {
	passwordSize := len(password)
	if passwordSize != 16 && passwordSize != 20 {
		return false, fmt.Errorf("expected value is either 16 or 20")
	}
	if err := c.connect(); err != nil {
		return false, err
	}
	uid, _, err := userIDs(user)
	if err != nil {
		return false, err
	}

	if _, err := c.client.TestUserPassword(defaultContext, uid, password, passwordSize == 20); err != nil {
		if respErr, ok := types.IsResponseError(err); ok {
			switch respErr.CompletionCode() {
			case types.CodeSetUserPasswordDataMismatch, types.CodeSetUserPasswordWrongSize:
				return true, fmt.Errorf("password for user %s with id %s incorrect: %w change necessary", user.Name, user.Id, err)
			}
		}
		return false, fmt.Errorf("error while testing user password for user %s with id %s: %w", user.Name, user.Id, err)
	}
	return false, nil
}

func (c *Client) changePassword(user api.BMCUser, passwd string, pc *api.PasswordConstraints) (string, error) {
	if err := c.connect(); err != nil {
		return "", err
	}
	uid, _, err := userIDs(user)
	if err != nil {
		return "", err
	}

	pw, err := c.setPassword(user, uid, passwd, pc)
	if err != nil {
		return "", err
	}

	if err := c.setUserEnabled(user, true); err != nil {
		return "", err
	}

	return pw, nil
}

func (c *Client) setPassword(user api.BMCUser, uid uint8, passwd string, pc *api.PasswordConstraints) (string, error) {
	err := retry.Do(
		func() error {
			pwd := passwd
			if pwd == "" && pc != nil {
				gen, err := password.Generate(pc.Length, pc.NumDigits, pc.NumSymbols, pc.NoUpper, pc.AllowRepeat)
				if err != nil {
					return fmt.Errorf("password generation failed for user:%s id:%s %w", user.Name, user.Id, err)
				}
				pwd = gen
			}
			if _, err := c.client.SetUserPassword(defaultContext, uid, pwd, len(pwd) > 16); err != nil {
				return fmt.Errorf("ipmi password creation failed for user:%s id:%s %w", user.Name, user.Id, err)
			}
			passwd = pwd
			return nil
		},
		retry.OnRetry(func(n uint, err error) {
			c.log.Infow("retry ipmi password creation", "user", user.Name, "id", user.Id, "retry", n, "cause", err)
		}),
		retry.Delay(1*time.Second),
		retry.Attempts(30),
	)
	return passwd, err
}

func (c *Client) setUserEnabled(user api.BMCUser, enabled bool) error {
	uid, _, err := userIDs(user)
	if err != nil {
		return err
	}

	if enabled {
		return retry.Do(
			func() error {
				if err := c.client.EnableUser(defaultContext, uid); err != nil {
					return fmt.Errorf("failed to enable user %s with id %s: %w", user.Name, user.Id, err)
				}
				return nil
			},
			retry.OnRetry(func(n uint, err error) {
				c.log.Infow("retry ipmi enable user", "user", user.Name, "id", user.Id, "retry", n, "cause", err)
			}),
			retry.Delay(1*time.Second),
			retry.Attempts(30),
		)
	}

	if err := c.client.DisableUser(defaultContext, uid); err != nil {
		return fmt.Errorf("failed to disable user %s with id %s: %w", user.Name, user.Id, err)
	}
	return nil
}

// OpenConsole connects to the serial console and puts the in/out into a ssh stream
func (c *Client) OpenConsole(s ssh.Session) error {
	if err := c.connect(); err != nil {
		return err
	}
	if _, err := io.WriteString(s, "Exit with ~X.\n"); err != nil {
		return fmt.Errorf("failed to write to console %w", err)
	}
	return c.client.SOLActivate(defaultContext, s, s, nil)
}

func userIDs(user api.BMCUser) (uid uint8, cn uint8, err error) {
	id, err := strconv.Atoi(user.Id)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid uid of user %s: %s %w", user.Name, user.Id, err)
	}
	return uint8(id), uint8(user.ChannelNumber), nil // nolint:gosec
}
