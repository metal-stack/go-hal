package dmi

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	guuid "github.com/google/uuid"
)

const (
	dmiUUID   = "/sys/class/dmi/id/product_uuid"
	dmiSerial = "/sys/class/dmi/id/product_serial"
)

// MachineUUID calculates a unique uuid for this (hardware) machine
func MachineUUID() (string, error) {
	return machineUUID(ioutil.ReadFile)
}

func machineUUID(readFileFunc func(filename string) ([]byte, error)) (string, error) {
	if _, err := os.Stat(dmiUUID); !os.IsNotExist(err) {
		productUUID, err := readFileFunc(dmiUUID)
		if err != nil {
			return "", fmt.Errorf("error getting product_uuid: %w", err)
		} else {
			return strings.TrimSpace(string(productUUID)), nil
		}
	}

	if _, err := os.Stat(dmiSerial); !os.IsNotExist(err) {
		productSerial, err := readFileFunc(dmiSerial)
		if err != nil {
			return "", fmt.Errorf("error getting product_serial: %w", err)
		} else {
			productSerialBytes, err := guuid.FromBytes([]byte(fmt.Sprintf("%16s", string(productSerial))))
			if err != nil {
				return "", fmt.Errorf("error converting product_serial to uuid:%w", err)
			} else {
				return strings.TrimSpace(productSerialBytes.String()), nil
			}
		}
	}
	return "00000000-0000-0000-0000-000000000000", fmt.Errorf("no valid UUID found")
}
