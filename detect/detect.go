package detect

import (
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/api"
	"github.com/metal-stack/go-hal/internal/dmi"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/internal/vendors/lenovo"
	"github.com/metal-stack/go-hal/internal/vendors/supermicro"
)

// InBand will try to detect the the board vendor
func InBand() (*api.Board, error) {
	b, err := dmi.BoardInfo()
	if err != nil {
		return nil, err
	}
	b.Vendor = api.GuessVendor(b.VendorString)
	return b, nil
}

// ConnectInBand will detect the board and choose the correct inband hal implementation
func ConnectInBand() (hal.InBand, error) {
	b, err := InBand()
	if err != nil {
		return nil, err
	}
	switch b.Vendor {
	case api.VendorLenovo:
		return lenovo.InBand()
	case api.VendorSupermicro:
		return supermicro.InBand("sum")
	case api.VendorUnknown:
		return nil, api.ErrorUnknownVendor
	default:
		return nil, api.ErrorUnknownVendor
	}
}

// OutBand will try to detect the the board vendor
func OutBand(ip, user, password string) (*api.Board, error) {
	r, err := redfish.New("https://"+ip, user, password, true)
	if err != nil {
		return nil, err
	}
	b, err := r.BoardInfo()
	if err != nil {
		return nil, err
	}
	b.Vendor = api.GuessVendor(b.VendorString)
	return b, nil
}

// ConnectOutBand will detect the board and choose the correct inband hal implementation
func ConnectOutBand(ip, user, password string) (hal.OutBand, error) {
	b, err := OutBand(ip, user, password)
	if err != nil {
		return nil, err
	}
	switch b.Vendor {
	case api.VendorLenovo:
		return lenovo.OutBand(&ip, &user, &password)
	case api.VendorSupermicro:
		return supermicro.OutBand("sum", true, &ip, &user, &password)
	case api.VendorUnknown:
		return nil, api.ErrorUnknownVendor
	default:
		return nil, api.ErrorUnknownVendor
	}
}
