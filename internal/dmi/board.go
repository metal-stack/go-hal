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

	d.readValues(boardMap)

	if boardMap[boardVendor] == "" {
		return nil, fmt.Errorf("board vendor could not be detected")
	}
	if boardMap[boardName] == "" {
		return nil, fmt.Errorf("board name could not be detected")
	}
	if boardMap[boardSerial] == "" {
		return nil, fmt.Errorf("board serial could not be detected")
	}

	return &api.Board{
		VendorString: boardMap[boardVendor],
		Model:        boardMap[boardName],
		SerialNumber: boardMap[boardSerial],
		PartNumber:   boardMap[productSerial],
		BiosVersion:  boardMap[biosVersion],
	}, nil
}
