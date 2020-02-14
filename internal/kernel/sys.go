package kernel

import "os"

var (
	sysfirmware = "/sys/firmware/efi"
)

// Firmware returns either efi or bios, depending on the boot method.
func Firmware() string {
	_, err := os.Stat(sysfirmware)
	if os.IsNotExist(err) {
		return "bios"
	}
	return "efi"
}
