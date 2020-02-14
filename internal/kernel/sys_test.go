package kernel

import (
	"os"
	"testing"
)

func TestFirmware(t *testing.T) {
	sysfirmware = "/tmp/testefi"
	_, err := os.OpenFile(sysfirmware, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(sysfirmware)

	firmware := Firmware()
	if firmware != "efi" {
		t.Error("expected efi firmware but didn't get")
	}

	sysfirmware = "/tmp/testbios"
	firmware = Firmware()
	if firmware != "bios" {
		t.Error("expected bios firmware but didn't get")
	}
}
