package dmi

import (
	"fmt"
	"os"
	"strings"

	"github.com/metal-stack/go-hal/pkg/api"
)

const (
	boardVendor   = "/sys/class/dmi/id/board_vendor"
	boardName     = "/sys/class/dmi/id/board_name"
	boardSerial   = "/sys/class/dmi/id/board_serial"
	productSerial = "/sys/class/dmi/id/product_serial"
	productName   = "/sys/class/dmi/id/product_name"
	sysVendor     = "/sys/class/dmi/id/sys_vendor"
	biosVersion   = "/sys/class/dmi/id/bios_version"
)

// BoardInfo return raw dmi data of the board.
//
// The board-level DMI files are not populated on every machine (for example
// many virtual machines only expose the system and product level DMI data).
// Therefore the board vendor falls back to the system vendor and the board
// name falls back to the product name. All other fields are optional and are
// left empty if they are not present.
func BoardInfo() (*api.Board, error) {
	var (
		vendor     = ""
		name       = ""
		serial     = ""
		partNumber = ""
		version    = ""

		err error
	)

	if vendor, err = dmi(boardVendor); err != nil {
		if vendor, err = dmi(sysVendor); err != nil {
			return nil, err
		}
	}

	if name, err = dmi(boardName); err != nil {
		name, _ = dmi(productName)
	}
	if serial, err = dmi(boardSerial); err != nil {
		serial = ""
	}
	if partNumber, err = dmi(productSerial); err != nil {
		partNumber = ""
	}
	if version, err = dmi(biosVersion); err != nil {
		version = ""
	}

	return &api.Board{
		VendorString: vendor,
		Model:        name,
		SerialNumber: serial,
		PartNumber:   partNumber,
		BiosVersion:  version,
	}, nil
}

func dmi(path string) (string, error) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		content, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("error getting content of %s: %w", path, err)
		}
		return strings.TrimSpace(string(content)), nil
	}

	return "", fmt.Errorf("%s does not exist", path)
}
