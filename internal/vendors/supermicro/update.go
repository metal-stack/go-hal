package supermicro

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
)

// UpdateBIOS updates BIOS.
func (s *sum) UpdateBIOS(reader io.Reader) error {
	biosUpdate := "biosUpdate"
	err := writeUpdate(biosUpdate, reader)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(biosUpdate)
	}()

	return s.execute("-c", "UpdateBios", "--file", biosUpdate, "--reboot")
}

// UpdateBMC updates BMC.
func (s *sum) UpdateBMC(reader io.Reader) error {
	bmcUpdate := "bmcUpdate"
	err := writeUpdate(bmcUpdate, reader)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(bmcUpdate)
	}()

	return s.execute("-c", "UpdateBmc", "--file", bmcUpdate, "--reboot")
}

func writeUpdate(filename string, reader io.Reader) error {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, buf.Bytes(), 0644)
}
