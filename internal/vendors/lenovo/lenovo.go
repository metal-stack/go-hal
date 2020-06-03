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
		common     *common.Common
		compliance api.Compliance
		board      *api.Board
		ip         string
		user       string
		password   string
	}
)

const ipmiToolBin = "ipmitool"

var (
	// errorNotImplemented for all funcs which are not implemented yet
	errorNotImplemented = fmt.Errorf("not implemented yet")
)

// InBand creates an inband connection to a Lenovo server.
func InBand(board *api.Board) (hal.InBand, error) {
	i, err := ipmi.New(ipmiToolBin, api.IPMI2Compliance)
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

// OutBand creates an outband connection to a Lenovo server.
func OutBand(r *redfish.APIClient, board *api.Board, ip, user, password string, compliance api.Compliance) (hal.OutBand, error) {
	return &outBand{
		common:     common.New(r),
		compliance: compliance,
		board:      board,
		ip:         ip,
		user:       user,
		password:   password,
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

func (i *inBand) IdentifyLEDState(state hal.IdentifyLEDState) error {
	switch state {
	case hal.IdentifyLEDStateOn:
		return i.IdentifyLEDOn()
	case hal.IdentifyLEDStateOff:
		return i.IdentifyLEDOff()
	default:
		return fmt.Errorf("unknown identify LED state: %s", state)
	}
}

func (i *inBand) IdentifyLEDOn() error {
	return i.i.SendChassisIdentifyRaw("0x00", "0x01")
}

func (i *inBand) IdentifyLEDOff() error {
	return i.i.SendChassisIdentifyRaw("0x00", "0x00")
}

func (i *inBand) BootFrom(bootTarget hal.BootTarget) error {
	return i.i.SendBootOrderRaw(bootTarget)
}

func (i *inBand) Firmware() (hal.FirmwareMode, error) {
	return i.common.Firmware()
}

func (i *inBand) SetFirmware(hal.FirmwareMode) error {
	return errorNotImplemented //TODO
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
	u, err := o.common.MachineUUID()
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
	return o.common.PowerReset() // PowerOn is not supported
}

func (o *outBand) PowerOff() error {
	return errorNotImplemented // PowerOff is not supported
}

func (o *outBand) PowerReset() error {
	return o.common.PowerReset()
}

func (o *outBand) PowerCycle() error {
	return o.common.PowerReset() // PowerCycle is not supported
}

func (o *outBand) IdentifyLEDState(state hal.IdentifyLEDState) error {
	switch state {
	case hal.IdentifyLEDStateOn:
		return o.IdentifyLEDOn()
	case hal.IdentifyLEDStateOff:
		return o.IdentifyLEDOff()
	default:
		return fmt.Errorf("unknown identify LED state: %s", state)
	}
}

func (o *outBand) IdentifyLEDOn() error {
	return errorNotImplemented
}

func (o *outBand) IdentifyLEDOff() error {
	return errorNotImplemented
}

func (o *outBand) BootFrom(target hal.BootTarget) error {
	return o.common.SetBootOrder(target, api.VendorLenovo)
}

func (o *outBand) Describe() string {
	return "OutBand connected to Lenovo"
}

func (o *outBand) Connection() (string, string, string) {
	return o.ip, o.user, o.password
}
