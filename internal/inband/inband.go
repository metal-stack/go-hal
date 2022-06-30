package inband

import (
	"github.com/google/uuid"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/dmi"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/kernel"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/metal-stack/go-hal/pkg/logger"
)

type InBand struct {
	IpmiTool ipmi.IpmiTool
	board    *api.Board
	dmi      *dmi.DMI
}

func New(board *api.Board, inspectBMC bool, log logger.Logger) (*InBand, error) {
	i, err := ipmi.New(log)
	if err != nil {
		return nil, err
	}

	dmi := dmi.New(log)

	if inspectBMC {
		bmc, err := i.BMC()
		if err != nil {
			return nil, err
		}
		board.BMC = bmc
		board.BIOS, err = dmi.Bios()
		if err != nil {
			return nil, err
		}
		board.Firmware = kernel.Firmware()
	}

	return &InBand{
		IpmiTool: i,
		board:    board,
		dmi:      dmi,
	}, nil
}

func (ib *InBand) Board() *api.Board {
	return ib.board
}

func (ib *InBand) UUID() (*uuid.UUID, error) {
	u, err := ib.dmi.MachineUUID()
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
