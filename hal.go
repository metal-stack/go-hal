package hal

import (
	"strings"

	"github.com/google/uuid"
	"github.com/metal-stack/go-hal/pkg/api"
)

type (
	// PowerState state of the power of a server
	PowerState int
	// BootTarget defines the way the server should boot
	BootTarget int
	// IdentifyLEDState the state of the LED to identify the server
	IdentifyLEDState int
	// FirmwareMode the Firmware mode of the server, either Legacy, Dual or Uefi
	FirmwareMode int
)

const (
	// PowerUnknownState the server power state is not known
	PowerUnknownState PowerState = iota
	// PowerOnState the server is powered on
	PowerOnState
	// PowerOffState the server is powered off
	PowerOffState
)
const (
	// BootTargetPXE the server boots via PXE
	BootTargetPXE BootTarget = iota + 1
	// BootTargetDisk the server boots from disk
	BootTargetDisk
	// BootTargetBIOS the server boots into Bios
	BootTargetBIOS
)
const (
	// IdentifyLEDStateUnknown the LED is unknown
	IdentifyLEDStateUnknown IdentifyLEDState = iota
	// IdentifyLEDStateOn the LED is on
	IdentifyLEDStateOn
	// IdentifyLEDStateOff the LED is off
	IdentifyLEDStateOff
)
const (
	// FirmwareModeUnknown server is in unknown firmware state
	FirmwareModeUnknown FirmwareMode = iota
	// FirmwareModeLegacy or BIOS
	FirmwareModeLegacy
	// FirmwareModeUEFI the server boots in uefi mode
	FirmwareModeUEFI
)

var (
	powerStates = [...]string{
		PowerOnState:      "ON",
		PowerOffState:     "OFF",
		PowerUnknownState: "UNKNOWN",
	}
	bootTargets = [...]string{
		BootTargetPXE:  "PXE",
		BootTargetDisk: "DISK",
		BootTargetBIOS: "BIOS",
	}
	ledStates = [...]string{
		IdentifyLEDStateOn:      "ON",
		IdentifyLEDStateOff:     "OFF",
		IdentifyLEDStateUnknown: "UNKNOWN",
	}
	firmwareModes = [...]string{
		FirmwareModeLegacy:  "LEGACY",
		FirmwareModeUEFI:    "UEFI",
		FirmwareModeUnknown: "UNKNOWN",
	}
)

// Stringer
func (p PowerState) String() string       { return powerStates[p] }
func (b BootTarget) String() string       { return bootTargets[b] }
func (i IdentifyLEDState) String() string { return ledStates[i] }
func (f FirmwareMode) String() string     { return firmwareModes[f] }

// GuessPowerState try to figure out the power state of the server
func GuessPowerState(powerState string) PowerState {
	for i, p := range powerStates {
		if strings.Contains(strings.ToLower(p), strings.ToLower(powerState)) {
			return PowerState(i)
		}
	}
	return PowerUnknownState
}

// InBand get and set settings from the server via the inband interface.
type InBand interface {
	// Board return board information of the current connection
	Board() *api.Board

	// UUID get the machine UUID
	// current usage in metal-hammer
	UUID() (*uuid.UUID, error)

	// PowerOff set power state of the server to off
	PowerOff() error
	// PowerReset reset the power state of the server
	PowerReset() error
	// PowerCycle cycle the power state of the server
	PowerCycle() error

	// IdentifyLEDState get the identify LED state
	IdentifyLEDState(IdentifyLEDState) error
	// IdentifyLEDOn set the identify LED to on
	IdentifyLEDOn() error
	// IdentifyLEDOff set the identify LED to off
	IdentifyLEDOff() error

	// BootFrom set the boot order of the server to the specified target
	BootFrom(BootTarget) error

	// Firmware get the FirmwareMode of the server
	Firmware() (FirmwareMode, error)
	// SetFirmware set the FirmwareMode of the server
	SetFirmware(FirmwareMode) error

	// Describe print a basic information about this connection
	Describe() string

	// TODO add MachineFRU, BiosVersion, BMCVersion, BMC{IP, MAC, Interface}

	// BMC related calls
	// BMCUser returns the details of the preset metal bmc user
	BMCUser() BMCUser
	// BMCPresent returns true if the InBand Connection found a usable BMC device
	BMCPresent() bool
	// Creates the given BMC user and returns generated password
	BMCCreateUserAndPassword(user BMCUser, privilege api.IpmiPrivilege, constraints api.PasswordConstraints) (string, error)
	// Creates the given BMC user with the given password
	BMCCreateUser(user BMCUser, privilege api.IpmiPrivilege, password string) error
	// Changes the password of the given BMC user
	BMCChangePassword(user BMCUser, newPassword string) error
	// Enables/Disables the given BMC user
	BMCSetUserEnabled(user BMCUser, enabled bool) error

	// ConfigureBIOS configures the BIOS regarding certain required options.
	// It returns whether the system needs to be rebooted afterwards
	ConfigureBIOS() (bool, error)

	// EnsureBootOrder ensures the boot order
	EnsureBootOrder(bootloaderID string) error
}

type BMCUser struct {
	Name          string
	Id            string
	ChannelNumber int
}

// OutBand get and set settings from the server via the out of band interface.
type OutBand interface {
	// Board return board information of the current connection
	Board() *api.Board
	// UUID get the machine uuid
	// current usage in ipmi-catcher
	UUID() (*uuid.UUID, error)

	// PowerState returns the power state of the server
	PowerState() (PowerState, error)
	// PowerOff set power state of the server to off
	PowerOff() error
	// PowerOn set power state of the server to on
	PowerOn() error
	// PowerReset reset the power state of the server
	PowerReset() error
	// PowerCycle cycle the power state of the server
	PowerCycle() error

	// IdentifyLEDState get the identify LED state
	IdentifyLEDState(IdentifyLEDState) error
	// IdentifyLEDOn set the identify LED to on
	IdentifyLEDOn() error
	// IdentifyLEDOff set the identify LED to off
	IdentifyLEDOff() error

	// BootFrom set the boot order of the server to the specified target
	BootFrom(BootTarget) error

	// Describe print a basic information about this connection
	Describe() string

	IPMIConnection() (ip string, port int, user, password string)

	// TODO implement console access from bmc-proxy
}
