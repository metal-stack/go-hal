package api

import (
	"fmt"
	"strings"

	"github.com/metal-stack/go-hal/internal/kernel"
)

// PasswordConstraints holds the constraints that are ensured for generated passwords
type PasswordConstraints struct {
	Length      int
	NumDigits   int
	NumSymbols  int
	NoUpper     bool
	AllowRepeat bool
}

// Privilege of an IPMI user
type IpmiPrivilege = uint8

const (
	// Callback IPMI privilege
	CallbackPrivilege IpmiPrivilege = iota + 1
	// User IPMI privilege
	UserPrivilege
	// Operator IPMI privilege
	OperatorPrivilege
	// Administrator IPMI privilege
	AdministratorPrivilege
	// OEM IPMI privilege
	OEMPrivilege
	// NoAccess IPMI privilege
	NoAccessPrivilege
)

// Board raw dmi board information
type Board struct {
	VM            bool
	VendorString  string
	Vendor        Vendor
	Model         string
	PartNumber    string
	SerialNumber  string
	BiosVersion   string
	BMC           *BMC
	BIOS          *BIOS
	Firmware      kernel.FirmwareMode
	IndicatorLED  string
	PowerMetric   *PowerMetric
	PowerSupplies []PowerSupply
}

type PowerMetric struct {
	// AverageConsumedWatts shall represent the
	// average power level that occurred averaged over the last IntervalInMin
	// minutes.
	AverageConsumedWatts float32
	// IntervalInMin shall represent the time
	// interval (or window), in minutes, in which the PowerMetrics properties
	// are measured over.
	// Should be an integer, but some Dell implementations return as a float.
	IntervalInMin float32
	// MaxConsumedWatts shall represent the
	// maximum power level in watts that occurred within the last
	// IntervalInMin minutes.
	MaxConsumedWatts float32
	// MinConsumedWatts shall represent the
	// minimum power level in watts that occurred within the last
	// IntervalInMin minutes.
	MinConsumedWatts float32
}

type PowerSupply struct {
	// Status shall contain any status or health properties
	// of the resource.
	Status Status
}

type Status struct {
	Health string
	State  string
}

// BMCUser holds BMC user details
type BMCUser struct {
	Name          string
	Id            string
	ChannelNumber int
}

// BMCConnection offers methods to add/update BMC users and retrieve BMC details
type BMCConnection interface {
	// BMC returns the actual BMC details
	BMC() (*BMC, error)
	// PresentSuperUser returns the details of the already present bmc superuser
	PresentSuperUser() BMCUser
	// NeedsPasswordChange checks if a password change is required
	NeedsPasswordChange(user BMCUser, password string) (bool, error)
	// SuperUser returns the details of the preset metal bmc superuser
	SuperUser() BMCUser
	// User returns the details of the preset metal bmc user
	User() BMCUser
	// Present returns true if the InBand Connection found a usable BMC device
	Present() bool
	// Creates the given BMC user and returns generated password
	CreateUserAndPassword(user BMCUser, privilege IpmiPrivilege) (string, error)
	// Creates the given BMC user with the given password
	CreateUser(user BMCUser, privilege IpmiPrivilege, password string) error
	// Changes the password of the given BMC user
	ChangePassword(user BMCUser, newPassword string) error
	// Enables/Disables the given BMC user
	SetUserEnabled(user BMCUser, enabled bool) error
}

// OutBandBMCConnection offers a method to retrieve BMC details
type OutBandBMCConnection interface {
	// BMC returns the actual BMC details
	BMC() (*BMC, error)
}

// BMC Base Management Controller details
type BMC struct {
	IP                  string
	MAC                 string
	ChassisPartNumber   string
	ChassisPartSerial   string
	BoardMfg            string
	BoardMfgSerial      string
	BoardPartNumber     string
	ProductManufacturer string
	ProductPartNumber   string
	ProductSerial       string
	FirmwareRevision    string `ipmitool:"Firmware Revision"`
}

// BIOS information of this machine
type BIOS struct {
	Version string
	Vendor  string
	Date    string
}

var (
	VagrantBoard = &Board{
		VM:           true,
		VendorString: "vagrant",
		Vendor:       VendorVagrant,
		Model:        "vagrant",
		PartNumber:   "vagrant",
		SerialNumber: "vagrant",
		BiosVersion:  "0",
		BMC: &BMC{
			IP:                  "1.1.1.1",
			MAC:                 "aa:bb:cc:dd:ee:ff",
			ChassisPartNumber:   "vagrant",
			ChassisPartSerial:   "vagrant",
			BoardMfg:            "vagrant",
			BoardMfgSerial:      "vagrant",
			BoardPartNumber:     "vagrant",
			ProductManufacturer: "vagrant",
			ProductPartNumber:   "vagrant",
			ProductSerial:       "vagrant",
			FirmwareRevision:    "vagrant",
		},
		BIOS: &BIOS{
			Version: "0",
			Vendor:  "vagrant",
			Date:    "01/01/2020",
		},
		Firmware: 0,
	}
)

type (
	// Vendor identifies different server vendors
	Vendor int
)

func (v Vendor) PasswordConstraints() *PasswordConstraints {
	return &PasswordConstraints{
		Length:      10,
		NumDigits:   3,
		NumSymbols:  0,
		NoUpper:     false,
		AllowRepeat: false,
	}
}

const (
	// VendorUnknown is a unknown Vendor
	VendorUnknown Vendor = iota
	// VendorSupermicro identifies all Supermicro servers
	VendorSupermicro
	// VendorNovarion identifies all Novarion servers
	VendorNovarion
	// VendorDell identifies all Dell servers
	VendorDell
	// VendorVagrant is a virtual machine.
	VendorVagrant
	// VendorGigabyte identifies all Gigabyte servers
	VendorGigabyte
)

var (
	vendors = [...]string{
		VendorSupermicro: "Supermicro",
		VendorNovarion:   "Novarion-Systems",
		VendorDell:       "Dell",
		VendorVagrant:    "Vagrant",
		VendorUnknown:    "UNKNOWN",
		VendorGigabyte:   "Giga Computing",
	}
	allVendors = [...]Vendor{VendorSupermicro, VendorNovarion, VendorDell, VendorVagrant, VendorUnknown, VendorGigabyte}
)

func (v Vendor) String() string { return vendors[v] }

// GuessVendor will try to guess from vendor string
func GuessVendor(vendor string) Vendor {
	for _, v := range allVendors {
		givenVendor := strings.TrimSpace(strings.ToLower(vendor))
		possibleVendor := strings.TrimSpace(strings.ToLower(v.String()))
		if strings.Contains(givenVendor, possibleVendor) {
			return v
		}
	}
	return VendorUnknown
}

func (b *Board) String() string {
	return fmt.Sprintf("Vendor:%s Model:%s", b.Vendor, b.Model)
}

func (b *BIOS) String() string {
	return "version:" + b.Version + " vendor:" + b.Vendor + " date:" + b.Date
}
