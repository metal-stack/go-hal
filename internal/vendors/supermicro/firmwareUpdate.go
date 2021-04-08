package supermicro

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
)

// UpdateBIOS updates given BIOS
func (s *sum) UpdateBIOS(reader io.Reader) error {
	return s.updateFirmware(reader, "UpdateBios")
}

// UpdateBMC updates given BMC
func (s *sum) UpdateBMC(reader io.Reader) error {
	return s.updateFirmware(reader, "UpdateBmc")
}

// updateFirmware updates given firmware
func (s *sum) updateFirmware(reader io.Reader, command string) error {
	firmwareUpdate, err := writeFirmwareUpdate(reader)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(firmwareUpdate)
	}()

	return s.execute("-c", command, "--file", firmwareUpdate, "--reboot")
}

func writeFirmwareUpdate(reader io.Reader) (string, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return "", err
	}

	tmp, err := ioutil.TempFile(".", "firmware.update-")
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
