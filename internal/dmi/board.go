package dmi

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/metal-stack/go-hal/internal/api"
)

const (
	boardVendor = "/sys/class/dmi/id/board_vendor"
	boardName   = "/sys/class/dmi/id/board_name"
)

// BoardInfo return raw dmi data of the board
func BoardInfo() (*api.Board, error) {
	vendor, err := dmi(boardVendor)
	if err != nil {
		return nil, err
	}
	name, err := dmi(boardName)
	if err != nil {
		return nil, err
	}
	return &api.Board{Vendor: vendor, Name: name}, nil
}

func dmi(path string) (string, error) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("error getting content of %s: %w", path, err)
		}
		return strings.TrimSpace(string(content)), nil
	}
	return "", fmt.Errorf("%s does not exist", path)
}
