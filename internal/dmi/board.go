package dmi

import (
	"github.com/metal-stack/go-hal/pkg/api"
)

const (
	boardVendor   = "/sys/class/dmi/id/board_vendor"
	boardName     = "/sys/class/dmi/id/board_name"
	boardSerial   = "/sys/class/dmi/id/board_serial"
	productSerial = "/sys/class/dmi/id/product_serial"
	biosVersion   = "/sys/class/dmi/id/bios_version"
)

// BoardInfo return raw dmi data of the board
func (d *DMI) BoardInfo() *api.Board {
	boardMap := map[string]string{
		boardVendor:   "",
		boardName:     "",
		boardSerial:   "",
		productSerial: "",
		biosVersion:   "",
	}

	for k := range boardMap {
		value, err := d.read(k)
		if err != nil {
			d.log.Errorw("board info not found", "path", k, "error", err)
			continue
		}

		boardMap[k] = value
	}

	return &api.Board{
		VendorString: boardMap[boardVendor],
		Model:        boardMap[boardName],
		SerialNumber: boardMap[boardSerial],
		PartNumber:   boardMap[productSerial],
		BiosVersion:  boardMap[biosVersion],
	}
}
