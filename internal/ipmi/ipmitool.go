package ipmi

// https://www.intel.com/content/dam/www/public/us/en/documents/product-briefs/ipmi-second-gen-interface-spec-v2-rev1-1.pdf

import (
	"bufio"
	"fmt"
	"github.com/metal-stack/go-hal"
	"github.com/sethvargo/go-password/password"
	"log"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/pkg/errors"
)

// IpmiTool defines methods to interact with IPMI
type IpmiTool interface {
	DevicePresent() bool
	NewCommand(arg ...string) (*exec.Cmd, error)
	Run(arg ...string) (string, error)
	CreateUser(channelNumber int, username, uid string, privilege api.IpmiPrivilege, constraints api.PasswordConstraints) (password string, err error)
	CreateUserRaw(channelNumber int, username, uid string, privilege api.IpmiPrivilege, constraints api.PasswordConstraints) (password string, err error)
	GetLanConfig() (LanConfig, error)
	SetBootOrder(target hal.BootTarget, vendor api.Vendor) error
	SetChassisControl(ChassisControlFunction) error
	SetChassisIdentifyLEDState(hal.IdentifyLEDState) error
	SetChassisIdentifyLEDOn() error
	SetChassisIdentifyLEDOff() error
	GetFru() (Fru, error)
	GetSession() (Session, error)
	BMC() (*api.BMC, error)
}

// Ipmitool is used to query and modify the IPMI based BMC from the host os
type Ipmitool struct {
	command string
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
func New() (IpmiTool, error) {
	ipmitoolBin := "ipmitool"
	_, err := exec.LookPath(ipmitoolBin)
	if err != nil {
		return nil, fmt.Errorf("ipmitool binary not present at:%s err:%w", ipmitoolBin, err)
	}
	return &Ipmitool{
		command: ipmitoolBin,
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
		return nil, errors.Wrapf(err, "unable to locate program:%s in path", i.command)
	}
	return exec.Command(path, args...), nil
}

// Run executes ipmitool with given arguments and returns the outcome
func (i *Ipmitool) Run(args ...string) (string, error) {
	cmd, err := i.NewCommand(args...)
	if err != nil {
		return "", err
	}
	output, err := cmd.Output()
	if err != nil {
		log.Printf("run ipmitool with args: %v output:%v error:%v", args, string(output), err)
	}
	return string(output), err
}

// GetFru returns the Field Replaceable Unit information
func (i *Ipmitool) GetFru() (Fru, error) {
	config := &Fru{}
	cmdOutput, err := i.Run("fru")
	if err != nil {
		return *config, errors.Wrapf(err, "unable to execute ipmitool 'fru':%v", cmdOutput)
	}
	fruMap := output2Map(cmdOutput)
	from(config, fruMap)
	return *config, nil
}

// GetBMCInfo returns the BMC info
func (i *Ipmitool) GetBMCInfo() (BMCInfo, error) {
	bmc := &BMCInfo{}
	cmdOutput, err := i.Run("bmc", "info")
	if err != nil {
		return *bmc, errors.Wrapf(err, "unable to execute ipmitool 'bmc info':%v", cmdOutput)
	}
	bmcMap := output2Map(cmdOutput)
	from(bmc, bmcMap)
	return *bmc, nil
}

// GetLanConfig returns the LAN config
func (i *Ipmitool) GetLanConfig() (LanConfig, error) {
	config := &LanConfig{}
	cmdOutput, err := i.Run("lan", "print")
	if err != nil {
		return *config, errors.Wrapf(err, "unable to execute ipmitool 'lan print':%v", cmdOutput)
	}
	lanConfigMap := output2Map(cmdOutput)
	from(config, lanConfigMap)
	return *config, nil
}

// GetSession returns the session
func (i *Ipmitool) GetSession() (Session, error) {
	session := &Session{}
	cmdOutput, err := i.Run("session", "info", "all")
	if err != nil {
		return *session, errors.Wrapf(err, "unable to execute ipmitool 'session info all':%v", cmdOutput)
	}
	sessionMap := output2Map(cmdOutput)
	from(session, sessionMap)
	return *session, nil
}

type createUserRequest struct {
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

// CreateUser creates an IPMI user with a generated password and given privilege level
func (i *Ipmitool) CreateUser(channelNumber int, username, uid string, privilege api.IpmiPrivilege, pwc api.PasswordConstraints) (string, error) {
	cn := strconv.Itoa(channelNumber)
	return i.createUser(createUserRequest{
		username:                   username,
		uid:                        uid,
		privilege:                  privilege,
		disableUserArgs:            []string{"user", "disable", uid},
		enableUserArgs:             []string{"user", "enable", uid},
		setUsernameArgs:            []string{"user", "set", "name", uid, username},
		setUserPrivilegeArgs:       []string{"channel", "setaccess", cn, uid, "link=on", "ipmi=on", "callin=on", fmt.Sprintf("privilege=%d", privilege)},
		enableSOLPayloadAccessArgs: []string{"sol", "payload", "enable", cn, uid},
		setPasswordFunc: func() (string, error) {
			return i.createPassword(username, uid, pwc)
		},
	})
}

// CreateUserRaw creates an IPMI user with a generated password and given privilege level through raw commands
func (i *Ipmitool) CreateUserRaw(channelNumber int, username, uid string, privilege api.IpmiPrivilege, pwc api.PasswordConstraints) (string, error) {
	id, err := strconv.Atoi(uid)
	if err != nil {
		return "", errors.Wrapf(err, "invalid uid of user %s: %s", username, uid)
	}
	userID := uint8(id)
	cn := uint8(channelNumber)

	return i.createUser(createUserRequest{
		username:                   username,
		uid:                        uid,
		privilege:                  privilege,
		disableUserArgs:            RawDisableUser(userID),
		enableUserArgs:             RawEnableUser(userID),
		setUsernameArgs:            RawSetUserName(userID, username),
		setUserPrivilegeArgs:       RawUserAccess(cn, userID, privilege),
		enableSOLPayloadAccessArgs: RawEnableUserSOLPayloadAccess(cn, userID),
		setPasswordFunc: func() (string, error) {
			return i.createPasswordRaw(username, userID, pwc)
		},
	})
}

func (i *Ipmitool) createUser(args createUserRequest) (string, error) {
	out, err := i.Run(args.setUsernameArgs...)
	if err != nil {
		return "", errors.Wrapf(err, "failed set username for user %s with id %s: %s", args.username, args.uid, out)
	}

	out, err = i.Run(args.disableUserArgs...)
	if err != nil {
		return "", errors.Wrapf(err, "failed to disable user with id %s: %s", args.uid, out)
	}

	pw, err := args.setPasswordFunc()
	if err != nil {
		return "", errors.Wrapf(err, "failed to set password %s for user %s with id %s: %s", pw, args.username, args.uid, out)
	}

	out, err = i.Run(args.enableUserArgs...)
	if err != nil {
		return "", errors.Wrapf(err, "failed to enable user %s with id %s: %s", args.username, args.uid, out)
	}

	out, err = i.Run(args.setUserPrivilegeArgs...)
	if err != nil {
		return "", errors.Wrapf(err, "failed to set privilege %d for user %s with id %s: %s", args.privilege, args.username, args.uid, out)
	}

	out, err = i.Run(args.enableSOLPayloadAccessArgs...)
	if err != nil {
		return "", errors.Wrapf(err, "failed to set enable user SOL payload access for user %s with id %s: %s", args.username, args.uid, out)
	}

	return pw, nil
}

// CreatePassword generates, sets and returns a password with given constraints for given IPMI user
func (i *Ipmitool) createPassword(username, uid string, pwc api.PasswordConstraints) (string, error) {
	s := func(pw string) []string {
		return []string{"user", "set", "password", uid, pw}
	}
	return i.createPw(username, uid, pwc, s)
}

// CreatePasswordRaw generates, sets (via raw bytes) and returns a password with given constraints for given IPMI user
func (i *Ipmitool) createPasswordRaw(username string, uid uint8, pwc api.PasswordConstraints) (string, error) {
	s := func(pw string) []string {
		return RawSetUserPassword(uid, pw)
	}
	return i.createPw(username, strconv.Itoa(int(uid)), pwc, s)
}

func (i *Ipmitool) createPw(username, uid string, pwc api.PasswordConstraints, setPasswordArgs func(string) []string) (string, error) {
	var pw string
	err := retry.Do(
		func() error {
			p, err := password.Generate(pwc.Length, pwc.NumDigits, pwc.NumSymbols, pwc.NoUpper, pwc.AllowRepeat)
			if err != nil {
				return errors.Wrapf(err, "password generation failed for user:%s id:%s", username, uid)
			}
			out, err := i.Run(setPasswordArgs(p)...)
			if err != nil {
				return errors.Wrapf(err, "ipmi password creation failed for user:%s id:%s output:%s", username, uid, out)
			}
			pw = p
			return nil
		},
		retry.OnRetry(func(n uint, err error) {
			log.Printf("retry ipmi password creation for user:%s id:%s retry:%d cause:%v", username, uid, n, err)
		}),
		retry.Delay(1*time.Second),
		retry.Attempts(30),
	)
	return pw, err
}

// SetBootOrder persistently sets the boot order to given target
func (i *Ipmitool) SetBootOrder(target hal.BootTarget, vendor api.Vendor) error {
	out, err := i.Run(RawSetSystemBootOptions(target, vendor)...)
	if err != nil {
		return errors.Wrapf(err, "unable to persistently set boot order:%s out:%v", target, out)
	}
	return nil
}

// SetChassisControl executes the given chassis control function
func (i *Ipmitool) SetChassisControl(fn ChassisControlFunction) error {
	_, err := i.Run(RawChassisControl(fn)...)
	if err != nil {
		return errors.Wrapf(err, "unable to set chassis control function:%X", fn)
	}
	return nil
}

// SetChassisIdentifyLEDOn turns on the chassis identify LED
func (i *Ipmitool) SetChassisIdentifyLEDState(state hal.IdentifyLEDState) error {
	switch state {
	case hal.IdentifyLEDStateOn:
		return i.SetChassisIdentifyLEDOn()
	case hal.IdentifyLEDStateOff:
		return i.SetChassisIdentifyLEDOff()
	default:
		return fmt.Errorf("unknown identify LED state: %s", state)
	}
}

// SetChassisIdentifyLEDOn turns on the chassis identify LED
func (i *Ipmitool) SetChassisIdentifyLEDOn() error {
	_, err := i.Run(RawChassisIdentifyOn()...)
	if err != nil {
		return errors.Wrapf(err, "unable to turn on the chassis identify LED")
	}
	return nil
}

// SetChassisIdentifyLEDOff turns off the chassis identify LED
func (i *Ipmitool) SetChassisIdentifyLEDOff() error {
	_, err := i.Run(RawChassisIdentifyOff()...)
	if err != nil {
		return errors.Wrapf(err, "unable to turn off the chassis identify LED")
	}
	return nil
}

func output2Map(cmdOutput string) map[string]string {
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(cmdOutput))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}
		value := strings.TrimSpace(strings.Join(parts[1:], ""))
		result[key] = value
	}
	for k, v := range result {
		log.Printf("output key:%s value:%s", k, v)
	}
	return result
}

// from uses reflection to fill a struct based on the tags on it
func from(target interface{}, input map[string]string) {
	log.Printf("from target:%s input:%s", target, input)
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
		case api.VendorSupermicro:
			bootDevQualifier = supermicroHDQualifier
		default:
			bootDevQualifier = hdQualifier
		}
	case hal.BootTargetBIOS:
		uefiQualifier = nextBootUEFIQualifier
		bootDevQualifier = biosQualifier
	}

	return
}
