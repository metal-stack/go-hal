package vagrant

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/internal/vendors/common"
	"github.com/metal-stack/go-hal/pkg/api"
	goipmi "github.com/vmware/goipmi"
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

const (
	ipmiToolBin = "ipmitool"
)

var (
	// errorNotImplemented for all funcs which are not implemented yet
	errorNotImplemented = fmt.Errorf("not implemented yet")
)

// InBand creates an inband connection to a vagrant VM.
func InBand(board *api.Board, compliance api.Compliance) (hal.InBand, error) {
	i, err := ipmi.New(ipmiToolBin, compliance)
	if err != nil {
		return nil, err
	}

	return &inBand{
		common: common.New(nil),
		i:      i,
		board:  board,
	}, nil
}

// OutBand creates an outband connection to a vagrant VM.
func OutBand(r *redfish.APIClient, board *api.Board, ip, user, password string, compliance api.Compliance) (hal.OutBand, error) {
	return &outBand{
		common:     common.New(r),
		board:      board,
		ip:         ip,
		user:       user,
		password:   password,
		compliance: compliance,
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
	return i.i.SendChassisControlRaw(goipmi.ControlPowerDown)
}

func (i *inBand) PowerReset() error {
	return i.i.SendChassisControlRaw(goipmi.ControlPowerHardReset)
}

func (i *inBand) PowerCycle() error {
	return i.i.SendChassisControlRaw(goipmi.ControlPowerCycle)
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
	return "InBand connected to Vagrant"
}

func (i *inBand) BMCPresent() bool {
	return i.i.DevicePresent()
}

func (i *inBand) BMCCreateUser(username, uid string) (string, error) {
	return i.i.CreateUser(username, uid, ipmi.Administrator)
}

// OutBand
func (o *outBand) Board() *api.Board {
	return o.board
}

func (o *outBand) UUID() (*uuid.UUID, error) {
	u, err := o.common.APIClient.MachineUUID()
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
	return o.sendChassisControl(goipmi.ControlPowerUp)
}

func (o *outBand) PowerOff() error {
	return o.sendChassisControl(goipmi.ControlPowerDown)
}

func (o *outBand) PowerReset() error {
	return o.sendChassisControl(goipmi.ControlPowerHardReset)
}

func (o *outBand) PowerCycle() error {
	return o.sendChassisControl(goipmi.ControlPowerCycle)
}

func (o *outBand) sendChassisControl(chassisControl goipmi.ChassisControl) error {
	client, err := ipmi.OpenClientConnection(o.Connection())
	if err != nil {
		return err
	}
	return client.Control(chassisControl)
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
	return o.sendChassisIdentifyRaw(0x00, 0x01)
}

func (o *outBand) IdentifyLEDOff() error {
	return o.sendChassisIdentifyRaw(0x00, 0x00)
}

func (o *outBand) sendChassisIdentifyRaw(intervalOrOff, forceOn uint8) error {
	client, err := ipmi.OpenClientConnection(o.Connection())
	if err != nil {
		return err
	}
	return ipmi.SendChassisIdentifyRaw(client, intervalOrOff, forceOn)
}

func (o *outBand) BootFrom(bootTarget hal.BootTarget) error {
	client, err := ipmi.OpenClientConnection(o.Connection())
	if err != nil {
		return err
	}

	useProgress := true
	// set set-in-progress flag
	err = ipmi.SendSystemBootRaw(client, goipmi.BootParamSetInProgress, 0x01)
	if err != nil {
		useProgress = false
	}

	err = ipmi.SendSystemBootRaw(client, goipmi.BootParamInfoAck, 0x01, 0x01)
	if err != nil {
		if useProgress {
			// set-in-progress = set-complete
			_ = ipmi.SendSystemBootRaw(client, goipmi.BootParamSetInProgress, 0x00)
		}
		return err
	}

	uefiQualifier, bootDevQualifier := ipmi.GetBootOrderQualifiers(bootTarget, o.compliance)
	uq, err := ipmi.Uint8(uefiQualifier)
	if err != nil {
		return err
	}
	bdq, err := ipmi.Uint8(bootDevQualifier)
	if err != nil {
		return err
	}
	err = ipmi.SendSystemBootRaw(client, goipmi.BootParamBootFlags, uq, bdq, 0x00, 0x00, 0x00)
	if err == nil {
		if useProgress {
			// set-in-progress = commit-write
			_ = ipmi.SendSystemBootRaw(client, goipmi.BootParamSetInProgress, 0x02)
		}
	}

	if useProgress {
		// set-in-progress = set-complete
		_ = ipmi.SendSystemBootRaw(client, goipmi.BootParamSetInProgress, 0x00)
	}

	return err
}

func (o *outBand) Describe() string {
	return "OutBand connected to Vagrant"
}

func (o *outBand) Connection() (string, string, string) {
	return o.ip, o.user, o.password
}
