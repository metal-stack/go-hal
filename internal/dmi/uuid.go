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
	m := map[string]string{
		productUUID:   "",
		productSerial: "",
	}

	d.readValues(m)

	if m[productUUID] != "" && isUUID(m[productUUID]) {
		return m[productUUID], nil
	}

	if m[productSerial] != "" && isUUID(m[productSerial]) {
		return m[productSerial], nil
	}

	return "", ErrNoUUIDFound
}

func isUUID(s string) bool {
	_, err := guuid.Parse(s)
	return err == nil
}
