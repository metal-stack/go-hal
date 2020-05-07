package bios

import (
	"io/ioutil"
	"os"
	"strings"

	log "github.com/inconshreveable/log15"
	"github.com/metal-stack/go-hal/pkg/api"
)

const (
	biosVersion = "/sys/class/dmi/id/bios_version"
	biosVendor  = "/sys/class/dmi/id/bios_vendor"
	biosDate    = "/sys/class/dmi/id/bios_date"
)

// Bios read bios informations
func Bios() *api.BIOS {
	return &api.BIOS{
		Version: read(biosVersion),
		Vendor:  read(biosVendor),
		Date:    read(biosDate),
	}
}

func read(file string) string {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Error("error reading", "file", file, "error", err)
			return ""
		}
		return strings.TrimSpace(string(content))
	}
	return ""
}
