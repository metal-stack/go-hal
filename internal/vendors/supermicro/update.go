package supermicro

import (
	"bytes"
	"io"
	"io/ioutil"
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
	firmwareUpdate := "firmware.update"
	err := writeUpdate(firmwareUpdate, reader)
	if err != nil {
		return err
	}
	//defer func() {
	//	_ = os.Remove(firmwareUpdate)
	//}()

	return s.execute("-c", command, "--file", firmwareUpdate, "--reboot")
}

func writeUpdate(filename string, reader io.Reader) error {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, buf.Bytes(), 0600)
}
