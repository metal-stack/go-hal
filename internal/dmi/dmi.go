package dmi

import (
	"strings"

	"github.com/metal-stack/go-hal/pkg/logger"
	"github.com/spf13/afero"
)

const (
	boardVendor   = "/sys/class/dmi/id/board_vendor"
	boardName     = "/sys/class/dmi/id/board_name"
	boardSerial   = "/sys/class/dmi/id/board_serial"
	productSerial = "/sys/class/dmi/id/product_serial"
	biosVersion   = "/sys/class/dmi/id/bios_version"
	productUUID   = "/sys/class/dmi/id/product_uuid"
)

type DMI struct {
	log logger.Logger
	fs  afero.Fs
}

func New(log logger.Logger) *DMI {
	return &DMI{
		log: log,
		fs:  afero.NewOsFs(),
	}
}

func (d *DMI) readWithTrim(path string) (string, error) {
	content, err := afero.ReadFile(d.fs, path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(content)), nil
}
