package dmi

import (
	"fmt"

	"github.com/metal-stack/go-hal/pkg/api"
)

// Bios returns bios info from dmi
func (d *DMI) Bios() (*api.BIOS, error) {
	biosMap := map[string]string{
		biosDate:    "",
		biosVendor:  "",
		biosVersion: "",
	}

	for k := range biosMap {
		value, err := d.readWithTrim(k)
		if err != nil {
			d.log.Errorw("bios info not found", "path", k, "error", err)
			continue
		}

		biosMap[k] = value
	}

	if biosMap[biosVendor] == "" {
		return nil, fmt.Errorf("bios vendor could not be detected")
	}

	return &api.BIOS{
		Date:    biosMap[biosDate],
		Vendor:  biosMap[biosVendor],
		Version: biosMap[biosVersion],
	}, nil
}
