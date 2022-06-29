package dmi

import (
	"fmt"

	guuid "github.com/google/uuid"
)

var (
	ErrNoUUIDFound = fmt.Errorf("no valid UUID found")
)

// MachineUUID calculates a unique uuid for this (hardware) machine
func (d *DMI) MachineUUID() (string, error) {
	content, err := d.readWithTrim(productUUID)
	if err == nil && isUUID(content) {
		return content, nil
	}

	d.log.Debugw("unable to determine dmi uuid", "from", productUUID, "error", err)

	content, err = d.readWithTrim(productSerial)
	if err == nil && isUUID(content) {
		return content, nil
	}

	d.log.Debugw("unable to determine dmi uuid", "from", productSerial, "error", err)

	return "", ErrNoUUIDFound
}

func isUUID(s string) bool {
	_, err := guuid.Parse(s)
	return err == nil
}
