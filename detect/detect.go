package detect

import (
	"fmt"
	"strings"

	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/dmi"
	"github.com/metal-stack/go-hal/internal/vendors/lenovo"
	"github.com/metal-stack/go-hal/internal/vendors/supermicro"
)

type (
	// Vendor identifies different server vendors
	Vendor int
)

const (
	// VendorSupermicro identifies all Supermicro servers
	VendorSupermicro Vendor = iota
	// VendorLenovo identifies all Lenovo servers
	VendorLenovo
	// VendorDell identifies all Dell servers
	VendorDell
	// VendorVagrant is a virtual machine.
	VendorVagrant
	// VendorUnknown is a unknown Vendor
	VendorUnknown
)

var (
	vendors             = [...]string{"Supermicro", "Lenovo", "Dell", "Vagrant", "UNKNOWN"}
	allVendors          = [...]Vendor{VendorSupermicro, VendorLenovo, VendorDell, VendorVagrant, VendorUnknown}
	errorNotImplemented = fmt.Errorf("not implemented yet")
	errorUnknownVendor  = fmt.Errorf("vendor unknown")
)

func (v Vendor) String() string { return vendors[v] }

// Board is the server board
type Board struct {
	// Vendor is the vendor of the server board
	Vendor Vendor
	// Name is the name of the server board
	Name string
}

func (b *Board) String() string {
	return fmt.Sprintf("Vendor:%s Name:%s", b.Vendor, b.Name)
}

// InBand will try to detect the the board vendor
func InBand() (*Board, error) {
	b, err := dmi.BoardInfo()
	if err != nil {
		return nil, err
	}
	return &Board{Vendor: guess(b.Vendor), Name: b.Name}, nil
}

// ConnectInBand will detect the board and choose the correct inband hal implementation
func ConnectInBand() (hal.InBand, error) {
	b, err := InBand()
	if err != nil {
		return nil, err
	}
	switch b.Vendor {
	case VendorLenovo:
		return lenovo.InBand()
	case VendorSupermicro:
		return supermicro.InBand("sum")
	case VendorUnknown:
		return nil, errorUnknownVendor
	default:
		return nil, errorUnknownVendor
	}
}

// OutBand will try to detect the the board vendor
func OutBand(ip, user, password string) (*Board, error) {
	return nil, errorNotImplemented
}

// ConnectOutBand will detect the board and choose the correct inband hal implementation
func ConnectOutBand(ip, user, password string) (hal.OutBand, error) {
	return nil, errorNotImplemented
}

func guess(vendor string) Vendor {
	for _, v := range allVendors {
		if strings.Contains(strings.ToLower(v.String()), strings.ToLower(vendor)) {
			return v
		}
	}
	return VendorUnknown
}
