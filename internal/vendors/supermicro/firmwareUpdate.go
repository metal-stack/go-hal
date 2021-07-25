package supermicro

import (
	"bytes"
	"io"
	"os"
)

// UpdateBIOS updates given BIOS
func (s *sum) UpdateBIOS(reader io.Reader) error {
	return s.updateFirmware(reader, "UpdateBios", true)
}

// UpdateBMC updates given BMC
func (s *sum) UpdateBMC(reader io.Reader) error {
	return s.updateFirmware(reader, "UpdateBmc", false)
}

// updateFirmware updates given firmware
func (s *sum) updateFirmware(reader io.Reader, command string, reboot bool) error {
	firmwareUpdate, err := writeFirmwareUpdate(reader)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(firmwareUpdate)
	}()

	args := []string{"-c", command, "--file", firmwareUpdate}
	if reboot {
		args = append(args, "--reboot")
	}
	return s.execute(args...)
}

func writeFirmwareUpdate(reader io.Reader) (string, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return "", err
	}

	tmp, err := os.CreateTemp(".", "firmware.update-")
	if err != nil {
		return "", err
	}

	_, err = tmp.Write(buf.Bytes())
	if err != nil {
		return "", err
	}

	err = tmp.Close()
	if err != nil {
		return "", err
	}

	return tmp.Name(), nil
}
