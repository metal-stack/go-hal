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
	biosVersion   = "/sys/class/dmi/id/bios_version"
)

// BoardInfo return raw dmi data of the board
func BoardInfo() (*api.Board, error) {
	vendor, err := dmi(boardVendor)
	if err != nil {
		return nil, err
	}
	name, err := dmi(boardName)
	if err != nil {
		return nil, err
	}
	bserial, err := dmi(boardSerial)
	if err != nil {
		return nil, err
	}
	pserial, err := dmi(productSerial)
	if err != nil {
		return nil, err
	}
	version, err := dmi(biosVersion)
	if err != nil {
		return nil, err
	}
	return &api.Board{
		VendorString: vendor,
		Model:        name,
		SerialNumber: bserial,
		PartNumber:   pserial,
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
