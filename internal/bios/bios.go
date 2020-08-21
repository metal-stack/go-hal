package bios

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/metal-stack/go-hal/pkg/api"
)

const (
	biosVersion = "/sys/class/dmi/id/bios_version"
	biosVendor  = "/sys/class/dmi/id/bios_vendor"
	biosDate    = "/sys/class/dmi/id/bios_date"
)

// Bios read bios information
func Bios() (*api.BIOS, error) {
	// vendor is required to detect the machine
	vendor, err := read(biosVendor)
	if err != nil {
		return nil, err
	}
	// version and date might be unknown
	version, err := read(biosVersion)
	if err != nil {
		version = "UNKNOWN"
	}
	date, err := read(biosDate)
	if err != nil {
		date = "UNKNOWN"
	}
	return &api.BIOS{
		Version: version,
		Vendor:  vendor,
		Date:    date,
	}, nil
}

func read(file string) (string, error) {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(content)), nil
	}
	return "", fmt.Errorf("%s does not exist", file)
}
