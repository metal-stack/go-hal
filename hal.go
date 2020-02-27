package hal

import (
	"strings"

	"github.com/google/uuid"
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
	// BootTargetBios the server boots into Bios
	BootTargetBios
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
		BootTargetBios: "BIOS",
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

// Guesser
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
	// UUID get the machine UUID
	// current usage in metal-hammer
	UUID() (*uuid.UUID, error)

	// PowerOff set power state of the server to off
	PowerOff() error
	// PowerOff reset the power state of the server
	PowerReset() error
	// PowerCycle cycle the power state of the server
	PowerCycle() error

	// BootFrom set the boot order of the server to the specified target
	BootFrom(BootTarget) error

	// Firmware get the FirmwareMode of the server
	Firmware() (FirmwareMode, error)
	// SetFirmware set the FirmwareMode of the server
	SetFirmware(FirmwareMode) error

	// TODO add MachineFRU, BiosVersion, BMCVersion, BMC{IP, MAC, User, Password, Interface}
}

// OutBand get and set settings from the server via the out of band interface.
type OutBand interface {
	// UUID get the machine uuid
	// current usage in ipmi-catcher
	UUID() (*uuid.UUID, error)

	// PowerState returns the power state of the server
	PowerState() (PowerState, error)
	// PowerOn set power state of the server to on
	PowerOn() error
	// PowerOff set power state of the server to off
	PowerOff() error
	// PowerOff reset the power state of the server
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

	// TODO implement console access from bmc-proxy
}
