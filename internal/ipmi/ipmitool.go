package ipmi

// https://www.intel.com/content/dam/www/public/us/en/documents/product-briefs/ipmi-second-gen-interface-spec-v2-rev1-1.pdf

import (
	"bufio"
	"fmt"
	"github.com/metal-stack/go-hal"
	"log"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/metal-stack/go-hal/internal/password"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/pkg/errors"
)

// IpmiTool defines methods to interact with IPMI
type IpmiTool interface {
	DevicePresent() bool
	Run(arg ...string) (string, error)
	CreateUser(username, uid string, privilege api.IpmiPrivilege) (string, error)
	CreateUserRaw(username, uid string, privilege api.IpmiPrivilege) (string, error)
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
func New(ipmitoolBin string) (IpmiTool, error) {
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

// Run executes ipmitool with given arguments
func (i *Ipmitool) Run(args ...string) (string, error) {
	path, err := exec.LookPath(i.command)
	if err != nil {
		return "", errors.Wrapf(err, "unable to locate program:%s in path", i.command)
	}
	cmd := exec.Command(path, args...)
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

// CreateUser creates an IPMI user with a generated password and given privilege level
func (i *Ipmitool) CreateUser(username, uid string, privilege api.IpmiPrivilege) (string, error) {
	out, err := i.Run("user", "set", "name", uid, username)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create user %s: %v", username, out)
	}
	// This happens from time to time for unknown reason
	// retry password creation max 30 times with 1 second delay
	pw := ""
	err = retry.Do(
		func() error {
			pw = password.Generate(10)
			out, err = i.Run("user", "set", "password", uid, pw)
			if err != nil {
				log.Printf("ipmi password creation failed for user:%s output:%v", username, out)
			}
			return err
		},
		retry.OnRetry(func(n uint, err error) {
			log.Printf("retry ipmi password creation for user:%s id:%s retry:%d", username, uid, n)
		}),
		retry.Delay(1*time.Second),
		retry.Attempts(30),
	)
	if err != nil {
		return pw, errors.Wrapf(err, "unable to set password for user %s: %v", username, out)
	}

	channelnumber := "1"
	out, err = i.Run("channel", "setaccess", channelnumber, uid, "link=on", "ipmi=on", "callin=on", fmt.Sprintf("privilege=%d", int(privilege)))
	if err != nil {
		return pw, errors.Wrapf(err, "unable to set privilege for user %s: %v", username, out)
	}
	out, err = i.Run("user", "enable", uid)
	if err != nil {
		return pw, errors.Wrapf(err, "unable to enable user %s: %v", username, out)
	}
	out, err = i.Run("sol", "payload", "enable", channelnumber, uid)
	if err != nil {
		return pw, errors.Wrapf(err, "unable to enable user %s for sol access: %v", username, out)
	}

	return pw, nil
}

// CreateUserRaw creates an IPMI user with a generated password and given privilege level through raw commands
func (i *Ipmitool) CreateUserRaw(username, uid string, privilege api.IpmiPrivilege) (string, error) {
	var out []string
	id, err := strconv.Atoi(uid)
	if err != nil {
		return "", errors.Wrapf(err, "invalid userID:%s", uid)
	}
	userID := uint8(id)

	o, err := i.Run(RawSetUserName(userID, username)...)
	if err != nil {
		return "", err
	}
	out = append(out, o)

	o, err = i.Run(RawDisableUser(userID)...)
	if err != nil {
		return "", err
	}
	out = append(out, o)

	pw := password.Generate(10)
	o, err = i.Run(RawSetUserPassword(userID, pw)...)
	if err != nil {
		return "", err
	}
	out = append(out, o)

	o, err = i.Run(RawEnableUser(userID)...)
	if err != nil {
		return "", err
	}
	out = append(out, o)

	channelNumber := uint8(1)
	o, err = i.Run(RawUserAccess(channelNumber, userID, privilege)...)
	if err != nil {
		return "", err
	}
	out = append(out, o)

	o, err = i.Run(RawEnableUserSOLPayloadAccess(channelNumber, userID)...)
	if err != nil {
		return "", err
	}
	out = append(out, o)

	return strings.Join(out, "\n"), nil
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
