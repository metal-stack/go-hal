package dmi

import (
	"fmt"

	"github.com/metal-stack/go-hal/pkg/api"
)

// BoardInfo return raw dmi data of the board
func (d *DMI) BoardInfo() (*api.Board, error) {
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
			return nil, fmt.Errorf("board info not found at %q: %w", k, err)
		}

		boardMap[k] = value
	}

	return &api.Board{
		VendorString: boardMap[boardVendor],
		Model:        boardMap[boardName],
		SerialNumber: boardMap[boardSerial],
		PartNumber:   boardMap[productSerial],
		BiosVersion:  boardMap[biosVersion],
	}, nil
}
