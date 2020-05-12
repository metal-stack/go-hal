package kernel

import "os"

type FirmwareMode int

const (
	BIOS FirmwareMode = iota + 1
	EFI
)

var (
	sysFirmware = "/sys/firmware/efi"
)

// Firmware returns either EFI or BIOS, depending on the boot method.
func Firmware() FirmwareMode {
	_, err := os.Stat(sysFirmware)
	if os.IsNotExist(err) {
		return BIOS
	}
	return EFI
}
