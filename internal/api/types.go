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
	allVendors         = [...]Vendor{VendorSupermicro, VendorLenovo, VendorDell, VendorVagrant, VendorUnknown}
	ErrorUnknownVendor = fmt.Errorf("vendor unknown")
)

func (v Vendor) String() string { return vendors[v] }
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
