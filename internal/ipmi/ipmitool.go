package ipmi

// https://www.intel.com/content/dam/www/public/us/en/documents/product-briefs/ipmi-second-gen-interface-spec-v2-rev1-1.pdf

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/console"
	"github.com/metal-stack/go-hal/pkg/logger"

	"github.com/sethvargo/go-password/password"

	"github.com/avast/retry-go/v4"
	"github.com/gliderlabs/ssh"
	"github.com/metal-stack/go-hal/pkg/api"
)

type ApiType int

const (
	HighLevel ApiType = iota // use ipmitool commands
	LowLevel                 // use raw commands
)

// IpmiTool defines methods to interact with IPMI
type IpmiTool interface {
	DevicePresent() bool
	NewCommand(arg ...string) (*exec.Cmd, error)
	Run(arg ...string) (string, error)
	CreateUser(user api.BMCUser, privilege api.IpmiPrivilege, password string, constraints *api.PasswordConstraints, apiType ApiType) (pwd string, err error)
	NeedsPasswordChange(user api.BMCUser, password string) (b bool, e error)
	ChangePassword(user api.BMCUser, newPassword string, apiType ApiType) error
	SetUserEnabled(user api.BMCUser, enabled bool, apiType ApiType) error
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

// Ipmitool is used to query and modify the IPMI based BMC from the host os
type Ipmitool struct {
	command  string
	ip       string
	port     int
	user     string
	password string
	outband  bool
	log      logger.Logger
}

func (i *Ipmitool) NeedsPasswordChange(user api.BMCUser, password string) (bool, error) {
	passwordSize := len(password)
	if passwordSize != 16 && passwordSize != 20 {
		return false, fmt.Errorf("expected value is either 16 or 20")
	}

	output, err := i.Run("user", "test", user.Id, strconv.Itoa(passwordSize), password)
	if err != nil {
		if strings.Contains(output, "Failure: password incorrect") {
			return true, fmt.Errorf("password for user %s with id %s incorrect: %w change necessary", user.Name, user.Id, err)
		}
		return false, fmt.Errorf("error while testing user password for user %s with id %s: %w", user.Name, user.Id, err)
	}
	return false, nil
}

// LanConfig contains the config of IPMI.
// Tag must contain first column name of ipmitool lan print command output
// to get the second column value be parsed into the field
type LanConfig struct {
	IP  string `ipmitool:"IP Address"`
	Mac string `ipmitool:"MAC Address"`
}

func (l *LanConfig) String() string {
	return fmt.Sprintf("ip: %s mac:%s", l.IP, l.Mac)
}

// Session holds information about the current IPMI session
type Session struct {
	UserID    string `ipmitool:"user id"`
	Privilege string `ipmitool:"privilege level"`
}

// Fru contains Field Replaceable Unit information, retrieved with 'ipmitool fru'
type Fru struct {
	ChassisPartNumber   string `ipmitool:"Chassis Part Number"`
	ChassisPartSerial   string `ipmitool:"Chassis Serial"`
	BoardMfg            string `ipmitool:"Board Mfg"`
	BoardMfgSerial      string `ipmitool:"Board Mfg Serial"`
	BoardPartNumber     string `ipmitool:"Board Part Number"`
	ProductManufacturer string `ipmitool:"Product Manufacturer"`
	ProductPartNumber   string `ipmitool:"Product Part Number"`
	ProductSerial       string `ipmitool:"Product Serial"`
}

// BMCInfo contains the parsed output of 'ipmitool bmc info'
type BMCInfo struct {
	FirmwareRevision string `ipmitool:"Firmware Revision"`
}

// New creates a new IpmiTool with the default command
func New(log logger.Logger) (IpmiTool, error) {
	ipmitoolBin := "ipmitool"
	_, err := exec.LookPath(ipmitoolBin)
	if err != nil {
		return nil, fmt.Errorf("ipmitool binary not present at:%s err:%w", ipmitoolBin, err)
	}
	return &Ipmitool{
		command: ipmitoolBin,
		log:     log,
	}, nil
}

// NewOutBand creates a new IpmiTool with the default command
func NewOutBand(ip string, port int, user, password string, log logger.Logger) (IpmiTool, error) {
	ipmitoolBin := "ipmitool"
	_, err := exec.LookPath(ipmitoolBin)
	if err != nil {
		return nil, fmt.Errorf("ipmitool binary not present at:%s err:%w", ipmitoolBin, err)
	}
	return &Ipmitool{
		command:  ipmitoolBin,
		ip:       ip,
		port:     port,
		user:     user,
		password: password,
		outband:  true,
		log:      log,
	}, nil
}

// BMC returns the BMC struct
func (i *Ipmitool) BMC() (*api.BMC, error) {
	lan, err := i.GetLanConfig()
	if err != nil {
		return nil, err
	}
	fru, err := i.GetFru()
	if err != nil {
		return nil, err
	}
	info, err := i.GetBMCInfo()
	if err != nil {
		return nil, err
	}
	bmc := &api.BMC{
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
		FirmwareRevision:    info.FirmwareRevision,
	}
	return bmc, nil
}

// DevicePresent returns true if the IPMI device is present, which is required to talk to the BMC
func (i *Ipmitool) DevicePresent() bool {
	const ipmiDevicePrefix = "/dev/ipmi*"
	matches, err := filepath.Glob(ipmiDevicePrefix)
	if err != nil {
		return false
	}
	return len(matches) > 0
}

// NewCommand returns a new ipmitool command with the given arguments
func (i *Ipmitool) NewCommand(args ...string) (*exec.Cmd, error) {
	path, err := exec.LookPath(i.command)
	if err != nil {
		return nil, fmt.Errorf("unable to locate program:%s in path %w", i.command, err)
	}
	return exec.Command(path, args...), nil
}

// Run executes ipmitool with given arguments and returns the outcome
func (i *Ipmitool) Run(args ...string) (string, error) {
	if i.outband {
		err := os.Setenv("IPMITOOL_PASSWORD", i.password)
		if err != nil {
			return "", err
		}
		defer func() {
			_ = os.Unsetenv("IPMITOOL_PASSWORD")
		}()
		args = append([]string{"-I", "lanplus", "-H", i.ip, "-p", strconv.Itoa(i.port), "-U", i.user, "-E"}, args...)
	}
	cmd, err := i.NewCommand(args...)
	if err != nil {
		return "", err
	}
	output, err := cmd.Output()
	if err != nil {
		i.log.Infow("run ipmitool", "args", args, "output", string(output), "error", err)
	}
	return string(output), err
}

// GetFru returns the Field Replaceable Unit information
func (i *Ipmitool) GetFru() (Fru, error) {
	config := &Fru{}
	cmdOutput, err := i.Run("fru")
	if err != nil {
		return *config, fmt.Errorf("unable to execute ipmitool 'fru':%v %w", cmdOutput, err)
	}
	fruMap := i.output2Map(cmdOutput)
	from(config, fruMap)
	return *config, nil
}

// GetBMCInfo returns the BMC info
func (i *Ipmitool) GetBMCInfo() (BMCInfo, error) {
	bmc := &BMCInfo{}
	cmdOutput, err := i.Run("bmc", "info")
	if err != nil {
		return *bmc, fmt.Errorf("unable to execute ipmitool 'bmc info':%v %w", cmdOutput, err)
	}
	bmcMap := i.output2Map(cmdOutput)
	from(bmc, bmcMap)
	return *bmc, nil
}

// GetLanConfig returns the LAN config
func (i *Ipmitool) GetLanConfig() (LanConfig, error) {
	config := &LanConfig{}
	cmdOutput, err := i.Run("lan", "print")
	if err != nil {
		return *config, fmt.Errorf("unable to execute ipmitool 'lan print':%v %w", cmdOutput, err)
	}
	lanConfigMap := i.output2Map(cmdOutput)
	from(config, lanConfigMap)
	return *config, nil
}

// GetSession returns the session
func (i *Ipmitool) GetSession() (Session, error) {
	session := &Session{}
	cmdOutput, err := i.Run("session", "info", "all")
	if err != nil {
		return *session, fmt.Errorf("unable to execute ipmitool 'session info all':%v %w", cmdOutput, err)
	}
	sessionMap := i.output2Map(cmdOutput)
	from(session, sessionMap)
	return *session, nil
}

type bmcRequest struct {
	username                   string
	uid                        string
	privilege                  api.IpmiPrivilege
	disableUserArgs            []string
	enableUserArgs             []string
	setUsernameArgs            []string
	setUserPrivilegeArgs       []string
	enableSOLPayloadAccessArgs []string
	setPasswordFunc            func() (string, error)
}

// CreateUser creates an IPMI user with given privilege level and either the given password or - if empty - a generated one with respect to the given password constraints
func (i *Ipmitool) CreateUser(user api.BMCUser, privilege api.IpmiPrivilege, password string, pc *api.PasswordConstraints, apiType ApiType) (string, error) {
	switch apiType {
	case LowLevel:
		id, err := strconv.Atoi(user.Id)
		if err != nil {
			return "", fmt.Errorf("invalid uid of user %s: %s %w", user.Name, user.Id, err)
		}
		userID := uint8(id)             // nolint:gosec
		cn := uint8(user.ChannelNumber) // nolint:gosec
		return i.createUser(bmcRequest{
			username:                   user.Name,
			uid:                        user.Id,
			privilege:                  privilege,
			disableUserArgs:            RawDisableUser(userID),
			enableUserArgs:             RawEnableUser(userID),
			setUsernameArgs:            RawSetUserName(userID, user.Name),
			setUserPrivilegeArgs:       RawUserAccess(cn, userID, privilege),
			enableSOLPayloadAccessArgs: RawEnableUserSOLPayloadAccess(cn, userID),
			setPasswordFunc: func() (string, error) {
				return i.createPasswordRaw(user.Name, userID, password, pc)
			},
		})
	case HighLevel:
		fallthrough
	default:
		cn := strconv.Itoa(user.ChannelNumber)
		return i.createUser(bmcRequest{
			username:                   user.Name,
			uid:                        user.Id,
			privilege:                  privilege,
			disableUserArgs:            []string{"user", "disable", user.Id},
			enableUserArgs:             []string{"user", "enable", user.Id},
			setUsernameArgs:            []string{"user", "set", "name", user.Id, user.Name},
			setUserPrivilegeArgs:       []string{"channel", "setaccess", cn, user.Id, "link=on", "ipmi=on", "callin=on", fmt.Sprintf("privilege=%d", privilege)},
			enableSOLPayloadAccessArgs: []string{"sol", "payload", "enable", cn, user.Id},
			setPasswordFunc: func() (string, error) {
				return i.createPassword(user.Name, user.Id, password, pc)
			},
		})
	}
}

// ChangePassword of the given user
func (i *Ipmitool) ChangePassword(user api.BMCUser, newPassword string, apiType ApiType) error {
	switch apiType {
	case LowLevel:
		id, err := strconv.Atoi(user.Id)
		if err != nil {
			return fmt.Errorf("invalid uid of user %s: %s %w", user.Name, user.Id, err)
		}
		userID := uint8(id) // nolint:gosec
		_, err = i.changePassword(bmcRequest{
			username:        user.Name,
			uid:             user.Id,
			disableUserArgs: RawDisableUser(userID),
			enableUserArgs:  RawEnableUser(userID),
			setPasswordFunc: func() (string, error) {
				return newPassword, nil
			},
		})
		return err
	case HighLevel:
		fallthrough
	default:
		_, err := i.changePassword(bmcRequest{
			username:        user.Name,
			uid:             user.Id,
			disableUserArgs: []string{"user", "disable", user.Id},
			enableUserArgs:  []string{"user", "enable", user.Id},
			setPasswordFunc: func() (string, error) {
				return newPassword, nil
			},
		})
		return err
	}
}

// SetUserEnabled enable the given user
func (i *Ipmitool) SetUserEnabled(user api.BMCUser, enabled bool, apiType ApiType) error {
	switch apiType {
	case LowLevel:
		id, err := strconv.Atoi(user.Id)
		if err != nil {
			return fmt.Errorf("invalid uid of user %s: %s %w", user.Name, user.Id, err)
		}
		userID := uint8(id) // nolint:gosec
		return i.setUserEnabled(bmcRequest{
			username:        user.Name,
			uid:             user.Id,
			disableUserArgs: RawDisableUser(userID),
			enableUserArgs:  RawEnableUser(userID),
		}, enabled)
	case HighLevel:
		fallthrough
	default:
		return i.setUserEnabled(bmcRequest{
			username:        user.Name,
			uid:             user.Id,
			disableUserArgs: []string{"user", "disable", user.Id},
			enableUserArgs:  []string{"user", "enable", user.Id},
		}, enabled)
	}
}

func (i *Ipmitool) createUser(req bmcRequest) (string, error) {
	out, err := i.Run(req.setUsernameArgs...)
	if err != nil {
		return "", fmt.Errorf("failed set username for user %s with id %s: %s %w", req.username, req.uid, out, err)
	}

	pw, err := i.changePassword(req)
	if err != nil {
		return "", err
	}

	out, err = i.Run(req.setUserPrivilegeArgs...)
	if err != nil {
		return "", fmt.Errorf("failed to set privilege %d for user %s with id %s: %s %w", req.privilege, req.username, req.uid, out, err)
	}

	out, err = i.Run(req.enableSOLPayloadAccessArgs...)
	if err != nil {
		return "", fmt.Errorf("failed to set enable user SOL payload access for user %s with id %s: %s %w", req.username, req.uid, out, err)
	}

	return pw, nil
}

func (i *Ipmitool) changePassword(req bmcRequest) (string, error) {
	pw, err := req.setPasswordFunc()
	if err != nil {
		return "", fmt.Errorf("failed to set password %s for user %s with id %s %w", pw, req.username, req.uid, err)
	}

	err = i.setUserEnabled(req, true)
	if err != nil {
		return "", err
	}

	return pw, nil
}

func (i *Ipmitool) setUserEnabled(req bmcRequest, enabled bool) error {
	if enabled {
		out, err := i.Run(req.enableUserArgs...)
		if err != nil {
			return fmt.Errorf("failed to enable user %s with id %s: %s %w", req.username, req.uid, out, err)
		}
		return nil
	}

	out, err := i.Run(req.disableUserArgs...)
	if err != nil {
		return fmt.Errorf("failed to disable user %s with id %s: %s %w", req.username, req.uid, out, err)
	}

	return nil
}

func (i *Ipmitool) createPassword(username, uid string, passwd string, pc *api.PasswordConstraints) (string, error) {
	s := func(pw string) []string {
		return []string{"user", "set", "password", uid, pw}
	}
	return i.createPw(username, uid, passwd, pc, s)
}

func (i *Ipmitool) createPasswordRaw(username string, uid uint8, passwd string, pc *api.PasswordConstraints) (string, error) {
	s := func(pw string) []string {
		return RawSetUserPassword(uid, pw)
	}
	return i.createPw(username, strconv.Itoa(int(uid)), passwd, pc, s)
}

func (i *Ipmitool) createPw(username, uid, passwd string, pc *api.PasswordConstraints, setPasswordArgs func(string) []string) (string, error) {
	err := retry.Do(
		func() error {
			pwd := passwd
			if pwd == "" && pc != nil {
				gen, err := password.Generate(pc.Length, pc.NumDigits, pc.NumSymbols, pc.NoUpper, pc.AllowRepeat)
				if err != nil {
					return fmt.Errorf("password generation failed for user:%s id:%s %w", username, uid, err)
				}
				pwd = gen
			}
			out, err := i.Run(setPasswordArgs(pwd)...)
			if err != nil {
				return fmt.Errorf("ipmi password creation failed for user:%s id:%s output:%s %w", username, uid, out, err)
			}
			passwd = pwd
			return nil
		},
		retry.OnRetry(func(n uint, err error) {
			i.log.Infow("retry ipmi password creation", "user", username, "id", uid, "retry", n, "cause", err)
		}),
		retry.Delay(1*time.Second),
		retry.Attempts(30),
	)
	return passwd, err
}

// SetBootOrder persistently sets the boot order to given target
func (i *Ipmitool) SetBootOrder(target hal.BootTarget, vendor api.Vendor) error {
	out, err := i.Run(RawSetSystemBootOptions(target, vendor)...)
	if err != nil {
		return fmt.Errorf("unable to persistently set boot order:%s out:%v %w", target, out, err)
	}
	return nil
}

// SetChassisControl executes the given chassis control function
func (i *Ipmitool) SetChassisControl(fn ChassisControlFunction) error {
	_, err := i.Run(RawChassisControl(fn)...)
	if err != nil {
		return fmt.Errorf("unable to set chassis control function:%X %w", fn, err)
	}
	return nil
}

// SetChassisIdentifyLEDState sets the chassis identify LED to given state
func (i *Ipmitool) SetChassisIdentifyLEDState(state hal.IdentifyLEDState) error {
	switch state {
	case hal.IdentifyLEDStateOn:
		return i.SetChassisIdentifyLEDOn()
	case hal.IdentifyLEDStateOff:
		return i.SetChassisIdentifyLEDOff()
	case hal.IdentifyLEDStateUnknown:
		fallthrough
	default:
		return fmt.Errorf("unknown identify LED state: %s", state)
	}
}

// SetChassisIdentifyLEDOn turns on the chassis identify LED
func (i *Ipmitool) SetChassisIdentifyLEDOn() error {
	_, err := i.Run(RawChassisIdentifyOn()...)
	if err != nil {
		return fmt.Errorf("unable to turn on the chassis identify LED %w", err)
	}
	return nil
}

// SetChassisIdentifyLEDOff turns off the chassis identify LED
func (i *Ipmitool) SetChassisIdentifyLEDOff() error {
	_, err := i.Run(RawChassisIdentifyOff()...)
	if err != nil {
		return fmt.Errorf("unable to turn off the chassis identify LED %w", err)
	}
	return nil
}

// OpenConsole connect to the serian console and put the in/out into a ssh stream
func (i *Ipmitool) OpenConsole(s ssh.Session) error {
	_, err := io.WriteString(s, "Exit with ~.\n")
	if err != nil {
		return fmt.Errorf("failed to write to console %w", err)
	}
	err = os.Setenv("IPMITOOL_PASSWORD", i.password)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Unsetenv("IPMITOOL_PASSWORD")
	}()
	cmd, err := i.NewCommand("-I", "lanplus", "-H", i.ip, "-p", strconv.Itoa(i.port), "-U", i.user, "-E", "sol", "activate")
	if err != nil {
		return err
	}
	return console.Open(s, cmd)
}

func (i *Ipmitool) output2Map(cmdOutput string) map[string]string {
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(cmdOutput))
	for scanner.Scan() {
		line := scanner.Text()
		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		value = strings.TrimSpace(value)
		result[key] = value
	}
	for k, v := range result {
		i.log.Debugw("output", "key", k, "value", v)
	}
	return result
}

// from uses reflection to fill a struct based on the tags on it
func from(target any, input map[string]string) {
	val := reflect.ValueOf(target).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		ipmitoolKey := tag.Get("ipmitool")
		valueField.SetString(input[ipmitoolKey])
	}
}

const (
	nextBootUEFIQualifier   = uint8(0xA0)
	persistentUEFIQualifier = uint8(0xE0)

	pxeQualifier = uint8(0x04)

	hdQualifier           = uint8(0x08)
	supermicroHDQualifier = uint8(0x24)

	biosQualifier = uint8(0x18)
)

// GetBootOrderQualifiers returns the qualifiers needed to set the given boot order according to the given vendor
func GetBootOrderQualifiers(bootTarget hal.BootTarget, vendor api.Vendor) (uefiQualifier, bootDevQualifier uint8) {
	switch bootTarget {
	case hal.BootTargetPXE:
		uefiQualifier = persistentUEFIQualifier
		bootDevQualifier = pxeQualifier
	case hal.BootTargetDisk:
		uefiQualifier = persistentUEFIQualifier
		switch vendor {
		case api.VendorSupermicro, api.VendorNovarion:
			bootDevQualifier = supermicroHDQualifier
		case api.VendorLenovo, api.VendorDell, api.VendorVagrant, api.VendorUnknown, api.VendorGigabyte:
			fallthrough
		default:
			bootDevQualifier = hdQualifier
		}
	case hal.BootTargetBIOS:
		uefiQualifier = nextBootUEFIQualifier
		bootDevQualifier = biosQualifier
	}

	return
}
