package dmi

import (
	"strings"

	"github.com/metal-stack/go-hal/pkg/logger"
	"github.com/spf13/afero"
)

const (
	boardVendor = "/sys/class/dmi/id/board_vendor"
	boardName   = "/sys/class/dmi/id/board_name"
	boardSerial = "/sys/class/dmi/id/board_serial"

	biosDate    = "/sys/class/dmi/id/bios_date"
	biosVendor  = "/sys/class/dmi/id/bios_vendor"
	biosVersion = "/sys/class/dmi/id/bios_version"

	productSerial = "/sys/class/dmi/id/product_serial"
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

func (d *DMI) readValues(m map[string]string) {
	for k := range m {
		content, err := afero.ReadFile(d.fs, k)
		if err != nil {
			d.log.Errorw("bios info not found", "path", k, "error", err)
			continue
		}

		m[k] = strings.TrimSpace(string(content))
	}
}
