package connect

import (
	"fmt"
	"github.com/metal-stack/go-hal/internal/vendors/vagrant"

	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/dmi"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/internal/vendors/lenovo"
	"github.com/metal-stack/go-hal/internal/vendors/supermicro"
	"github.com/metal-stack/go-hal/pkg/api"
)

var (
	errorUnknownVendor = fmt.Errorf("vendor unknown")
)

// InBand will detect the board and choose the correct inband hal implementation
func InBand() (hal.InBand, error) {
	b, err := dmi.BoardInfo()
	if err != nil {
		b = api.VagrantBoard
	}
	b.Vendor = api.GuessVendor(b.VendorString)

	switch b.Vendor {
	case api.VendorLenovo:
		return lenovo.InBand(b)
	case api.VendorSupermicro:
		return supermicro.InBand(b)
	case api.VendorVagrant:
		return vagrant.InBand(b)
	default:
		return nil, errorUnknownVendor
	}
}

// OutBand will detect the board and choose the correct outband hal implementation
func OutBand(ip string, ipmiPort int, user, password string) (hal.OutBand, error) {
	r, err := redfish.New("https://"+ip, user, password, true)
	if err != nil {
		return nil, err
	}
	b, err := r.BoardInfo()
	if err != nil {
		return nil, err
	}
	b.Vendor = api.GuessVendor(b.VendorString)

	switch b.Vendor {
	case api.VendorLenovo:
		return lenovo.OutBand(r, b), nil
	case api.VendorSupermicro:
		return supermicro.OutBand(r, b, ip, ipmiPort, user, password)
	case api.VendorVagrant:
		return vagrant.OutBand(b, ip, ipmiPort, user, password), nil
	default:
		return nil, errorUnknownVendor
	}
}
