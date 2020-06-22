package inband

import (
	"github.com/google/uuid"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/bios"
	"github.com/metal-stack/go-hal/internal/dmi"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/kernel"
	"github.com/metal-stack/go-hal/pkg/api"
)

type InBand struct {
	Ipmi  ipmi.Ipmi
	board *api.Board
}

func New(board *api.Board, inspectBMC bool) (*InBand, error) {
	i, err := ipmi.New("ipmitool")
	if err != nil {
		return nil, err
	}

	if inspectBMC {
		bmc, err := i.BMC()
		if err != nil {
			return nil, err
		}
		board.BMC = bmc
		board.BIOS = bios.Bios()
		board.Firmware = kernel.Firmware()
	}

	return &InBand{
		Ipmi:  i,
		board: board,
	}, nil
}

func (ib *InBand) Board() *api.Board {
	return ib.board
}

func (ib *InBand) UUID() (*uuid.UUID, error) {
	u, err := dmi.MachineUUID()
	if err != nil {
		return nil, err
	}
	us, err := uuid.Parse(u)
	if err != nil {
		return nil, err
	}
	return &us, nil
}

func (ib *InBand) Firmware() (hal.FirmwareMode, error) {
	var firmware hal.FirmwareMode
	switch kernel.Firmware() {
	case kernel.BIOS:
		firmware = hal.FirmwareModeLegacy
	case kernel.EFI:
		firmware = hal.FirmwareModeUEFI
	default:
		firmware = hal.FirmwareModeUnknown
	}
	return firmware, nil
}
