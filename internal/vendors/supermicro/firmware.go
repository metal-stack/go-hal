package supermicro

import (
	"bytes"
	"io"
	"os"
)

// UpdateBIOS updates given BIOS
func (s *sum) UpdateBIOS(reader io.Reader) error {
	return s.updateFirmware(reader, "UpdateBios", "--reboot", "--preserve_setting")
}

// UpdateBMC updates given BMC
func (s *sum) UpdateBMC(reader io.Reader) error {
	return s.updateFirmware(reader, "UpdateBmc")
}

// updateFirmware updates given firmware
func (s *sum) updateFirmware(reader io.Reader, command string, additionalArgs ...string) error {
	firmwareUpdate, err := writeFirmwareUpdate(reader)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(firmwareUpdate)
	}()

	args := []string{"-c", command, "--file", firmwareUpdate}
	args = append(args, additionalArgs...)

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
