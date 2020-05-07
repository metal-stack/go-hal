package api

import (
	"fmt"
	"strings"
)

// Board raw dmi board information
type Board struct {
	VendorString string
	Vendor       Vendor
	Model        string
	PartNumber   string
	SerialNumber string
	BiosVersion  string
	BMC          *BMC
	BIOS         *BIOS
}

// BMC Base Management Controller details
type BMC struct {
	IP  string
	MAC string

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

type (
	// Vendor identifies different server vendors
	Vendor int
)

const (
	// VendorUnknown is a unknown Vendor
	VendorUnknown Vendor = iota
	// VendorSupermicro identifies all Supermicro servers
	VendorSupermicro
	// VendorLenovo identifies all Lenovo servers
	VendorLenovo
	// VendorDell identifies all Dell servers
	VendorDell
	// VendorVagrant is a virtual machine.
	VendorVagrant
)

var (
	vendors = [...]string{
		VendorSupermicro: "Supermicro",
		VendorLenovo:     "Lenovo",
		VendorDell:       "Dell",
		VendorVagrant:    "Vagrant",
		VendorUnknown:    "UNKNOWN",
	}
	allVendors = [...]Vendor{VendorSupermicro, VendorLenovo, VendorDell, VendorVagrant, VendorUnknown}
)

func (v Vendor) String() string { return vendors[v] }

// GuessVendor will try to guess from vendor string
func GuessVendor(vendor string) Vendor {
	for _, v := range allVendors {
		if strings.Contains(strings.ToLower(v.String()), strings.ToLower(vendor)) {
			return v
		}
	}
	return VendorUnknown
}

func (b *Board) String() string {
	return fmt.Sprintf("Vendor:%s Name:%s", b.Vendor, b.Model)
}

func (b *BIOS) String() string {
	return "version:" + b.Version + " vendor:" + b.Vendor + " date:" + b.Date
}
