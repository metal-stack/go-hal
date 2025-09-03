package connect

import (
	"fmt"
	"github.com/metal-stack/go-hal/internal/vendors/gigabyte"

	"github.com/metal-stack/go-hal/internal/vendors/vagrant"
	"github.com/metal-stack/go-hal/pkg/logger"

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
func InBand(log logger.Logger) (hal.InBand, error) {
	b, err := dmi.BoardInfo()
	if err != nil {
		b = api.VagrantBoard
	}
	b.Vendor = api.GuessVendor(b.VendorString)
	log.Debugw("connect", "vendor", b)
	switch b.Vendor {
	case api.VendorLenovo:
		return lenovo.InBand(b, log)
	case api.VendorSupermicro, api.VendorNovarion:
		return supermicro.InBand(b, log)
	case api.VendorVagrant:
		return vagrant.InBand(b, log)
	case api.VendorGigabyte:
		return gigabyte.InBand(b, log)
	case api.VendorDell, api.VendorUnknown:
		fallthrough
	default:
		log.Errorw("connect", "unknown vendor", b.Vendor)
		return nil, errorUnknownVendor
	}
}

// OutBand will detect the board and choose the correct outband hal implementation
func OutBand(ip string, ipmiPort int, user, password string, log logger.Logger) (hal.OutBand, error) {
	r, err := redfish.New("https://"+ip, user, password, true, log)
	if err != nil {
		return nil, fmt.Errorf("Unable to establish redfish connection for ip:%s user:%s error:%w", ip, user, err)
	}
	b, err := r.BoardInfo()
	if err != nil {
		return nil, fmt.Errorf("Unable to get board info via redfish for ip:%s user:%s error:%w", ip, user, err)
	}
	b.Vendor = api.GuessVendor(b.VendorString)
	log.Debugw("connect", "board", b)
	switch b.Vendor {
	case api.VendorLenovo:
		return lenovo.OutBand(r, b), nil
	case api.VendorSupermicro, api.VendorNovarion:
		return supermicro.OutBand(r, b, ip, ipmiPort, user, password, log)
	case api.VendorVagrant:
		return vagrant.OutBand(b, ip, ipmiPort, user, password), nil
	case api.VendorGigabyte:
		return gigabyte.OutBand(r, b), nil
	case api.VendorDell, api.VendorUnknown:
		fallthrough
	default:
		log.Errorw("connect", "unknown vendor", b.Vendor)
		return nil, errorUnknownVendor
	}
}
