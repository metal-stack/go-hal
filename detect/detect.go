package detect

import (
	"fmt"

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

// inBand will try to detect the board vendor
func inBand() *api.Board {
	b, err := dmi.BoardInfo()
	if err != nil {
		b = &api.Board{
			VendorString: "vagrant",
			Vendor:       0,
			Model:        "vagrant",
			PartNumber:   "vagrant",
			SerialNumber: "vagrant",
			BiosVersion:  "0",
			BMC:          &api.BMC{
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
			BIOS:         &api.BIOS{
				Version: "0",
				Vendor:  "vagrant",
				Date: "01/01/2020",
			},
			Firmware:     0,
		}
	}
	b.Vendor = api.GuessVendor(b.VendorString)
	return b
}

// ConnectInBand will detect the board and choose the correct inband hal implementation
func ConnectInBand(compliance api.Compliance) (hal.InBand, error) {
	b := inBand()
	switch b.Vendor {
	case api.VendorLenovo:
		return lenovo.InBand(b, compliance)
	case api.VendorSupermicro:
		return supermicro.InBand(b, compliance)
	default:
		return nil, errorUnknownVendor
	}
}

// outBand will try to detect the board vendor
func outBand(r *redfish.APIClient) (*api.Board, error) {
	b, err := r.BoardInfo()
	if err != nil {
		return nil, err
	}
	b.Vendor = api.GuessVendor(b.VendorString)
	return b, nil
}

// ConnectOutBand will detect the board and choose the correct inband hal implementation
func ConnectOutBand(ip, user, password string, compliance api.Compliance) (hal.OutBand, error) {
	r, err := redfish.New("https://"+ip, user, password, true)
	if err != nil {
		return nil, err
	}
	b, err := outBand(r)
	if err != nil {
		return nil, err
	}
	switch b.Vendor {
	case api.VendorLenovo:
		return lenovo.OutBand(r, b, ip, user, password, compliance)
	case api.VendorSupermicro:
		return supermicro.OutBand(r, b, ip, user, password, compliance)
	default:
		return nil, errorUnknownVendor
	}
}
