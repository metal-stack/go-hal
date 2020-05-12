package kernel

import (
	"os"
	"testing"
)

func TestFirmware(t *testing.T) {
	sysFirmware = "/tmp/testefi"
	_, err := os.OpenFile(sysFirmware, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(sysFirmware)

	firmware := Firmware()
	if firmware != EFI {
		t.Error("expected efi firmware but didn't get")
	}

	sysFirmware = "/tmp/testbios"
	firmware = Firmware()
	if firmware != BIOS {
		t.Error("expected bios firmware but didn't get")
	}
}
