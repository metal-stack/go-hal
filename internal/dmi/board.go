package dmi

import (
	"github.com/metal-stack/go-hal/pkg/api"
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
		value, err := d.readWithTrim(k)
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
