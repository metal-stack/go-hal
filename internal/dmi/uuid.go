package dmi

import (
	"fmt"

	guuid "github.com/google/uuid"
)

var (
	ErrNoUUIDFound = fmt.Errorf("no valid UUID found")
)

const (
	dmiUUID   = "/sys/class/dmi/id/product_uuid"
	dmiSerial = "/sys/class/dmi/id/product_serial"
)

// MachineUUID calculates a unique uuid for this (hardware) machine
func (d *DMI) MachineUUID() (string, error) {
	_, err := d.fs.Stat(dmiUUID)
	if err == nil {
		return d.read(dmiUUID)
	}

	_, err = d.fs.Stat(dmiSerial)
	if err == nil {
		productSerial, err := d.read(dmiSerial)
		if err != nil {
			return "", err
		}

		_, err = guuid.Parse(productSerial)
		if err == nil {
			return productSerial, nil
		}
	}

	return "", ErrNoUUIDFound
}
