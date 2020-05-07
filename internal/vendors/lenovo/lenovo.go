package lenovo

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/bios"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/kernel"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/internal/vendors/common"
	"github.com/metal-stack/go-hal/pkg/api"
)

type (
	inBand struct {
		common *common.Common
		i      ipmi.Ipmi
		board  *api.Board
	}
	outBand struct {
		redfish *redfish.APIClient
		common  *common.Common
		board   *api.Board
	}
)

var (
	// errorNotImplemented for all funcs which are not implemented yet
	errorNotImplemented = fmt.Errorf("not implemented yet")
)

// InBand create a inband connection to a supermicro server.
func InBand(board *api.Board) (hal.InBand, error) {
	i, err := ipmi.New("ipmitool")
	if err != nil {
		return nil, err
	}
	bmc, err := i.BMC()
	if err != nil {
		return nil, err
	}
	board.BMC = bmc
	board.BIOS = bios.Bios()
	board.Firmware = kernel.Firmware()

	return &inBand{
		common: common.New(nil),
		i:      i,
		board:  board,
	}, nil
}

// OutBand create a outband connection to a supermicro server.
func OutBand(board *api.Board, ip, user, password *string) (hal.OutBand, error) {
	r, err := redfish.New("https://"+*ip, *user, *password, true)
	if err != nil {
		return nil, err
	}
	return &outBand{
		redfish: r,
		common:  common.New(r),
		board:   board,
	}, nil
}

// InBand

func (i *inBand) Board() *api.Board {
	return i.board
}
func (i *inBand) UUID() (*uuid.UUID, error) {
	return i.common.UUID()
}
func (i *inBand) PowerOff() error {
	return errorNotImplemented
}
func (i *inBand) PowerReset() error {
	return errorNotImplemented
}
func (i *inBand) PowerCycle() error {
	return errorNotImplemented
}
func (i *inBand) BootFrom(hal.BootTarget) error {
	return errorNotImplemented
}
func (i *inBand) Firmware() (hal.FirmwareMode, error) {
	return i.common.Firmware()
}
func (i *inBand) SetFirmware(hal.FirmwareMode) error {
	return errorNotImplemented
}
func (i *inBand) Describe() string {
	return "InBand connected to Lenovo"
}
func (i *inBand) BMCPresent() bool {
	return i.i.DevicePresent()
}
func (i *inBand) BMCCreateUser(username, uid string) (string, error) {
	return "", errorNotImplemented
}

// OutBand
func (o *outBand) Board() *api.Board {
	return o.board
}
func (o *outBand) UUID() (*uuid.UUID, error) {
	u, err := o.redfish.MachineUUID()
	if err != nil {
		return nil, err
	}
	us, err := uuid.Parse(u)
	if err != nil {
		return nil, err
	}
	return &us, nil
}
func (o *outBand) PowerState() (hal.PowerState, error) {
	return o.common.PowerState()
}
func (o *outBand) PowerOn() error {
	return errorNotImplemented
}
func (o *outBand) PowerOff() error {
	return errorNotImplemented
}

func (o *outBand) PowerReset() error {
	return errorNotImplemented
}

func (o *outBand) PowerCycle() error {
	return errorNotImplemented
}

func (o *outBand) IdentifyLEDState(hal.IdentifyLEDState) error {
	return errorNotImplemented
}

func (o *outBand) IdentifyLEDOn() error {
	return errorNotImplemented
}

func (o *outBand) IdentifyLEDOff() error {
	return errorNotImplemented
}

func (o *outBand) BootFrom(hal.BootTarget) error {
	return errorNotImplemented
}
func (o *outBand) Describe() string {
	return "OutBand connected to Lenovo"
}
