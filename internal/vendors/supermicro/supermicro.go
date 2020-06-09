package supermicro

import (
	"fmt"
	goipmi "github.com/vmware/goipmi"

	"github.com/google/uuid"
	hal "github.com/metal-stack/go-hal"
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
		sum    *sum
	}
	outBand struct {
		common     *common.Common
		compliance api.Compliance
		board      *api.Board
		sum        *sum
		ip         string
		user       string
		password   string
	}
)

const (
	sumBin      = "sum"
	ipmiToolBin = "ipmitool"
)

var (
	// errorNotImplemented for all funcs which are not implemented yet
	errorNotImplemented = fmt.Errorf("not implemented yet")
)

// InBand creates an inband connection to a supermicro server.
func InBand(board *api.Board) (hal.InBand, error) {
	s, err := newSum(sumBin)
	if err != nil {
		return nil, err
	}
	i, err := ipmi.New(ipmiToolBin, api.SMCIPMIToolCompliance)
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
		sum:    s,
	}, nil
}

// OutBand creates an outband connection to a supermicro server.
func OutBand(r *redfish.APIClient, board *api.Board, ip, user, password string, compliance api.Compliance) (hal.OutBand, error) {
	s, err := newRemoteSum(sumBin, ip, user, password)
	if err != nil {
		return nil, err
	}
	return &outBand{
		common:     common.New(r),
		board:      board,
		sum:        s,
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
	return i.i.ExecuteChassisControlFunction(ipmi.ChassisControlPowerUp)
}

func (i *inBand) PowerCycle() error {
	return i.i.ExecuteChassisControlFunction(ipmi.ChassisControlPowerCycle)
}

func (i *inBand) PowerReset() error {
	return i.i.ExecuteChassisControlFunction(ipmi.ChassisControlHardReset)
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
	return i.i.SetChassisIdentifyLEDOn()
}

func (i *inBand) IdentifyLEDOff() error {
	return i.i.SetChassisIdentifyLEDOff()
}

func (i *inBand) BootFrom(bootTarget hal.BootTarget) error {
	return i.i.SetBootOrder(bootTarget)
}

func (i *inBand) Firmware() (hal.FirmwareMode, error) {
	return i.common.Firmware()
}

func (i *inBand) SetFirmware(hal.FirmwareMode) error {
	return errorNotImplemented //TODO
}

func (i *inBand) Describe() string {
	return "InBand connected to Supermicro"
}

func (i *inBand) BMCPresent() bool {
	return i.i.DevicePresent()
}

func (i *inBand) BMCCreateUser(username, uid string) (string, error) {
	return i.i.CreateUser(username, uid, ipmi.AdministratorPrivilege)
}

// OutBand
func (o *outBand) Board() *api.Board {
	return o.board
}

func (o *outBand) UUID() (*uuid.UUID, error) {
	u, err := o.common.APIClient.MachineUUID()
	if err != nil {
		u, err = o.sum.uuidRemote()
		if err != nil {
			return nil, err
		}
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
	return o.setChassisControl(goipmi.ControlPowerUp)
}

func (o *outBand) PowerOff() error {
	return o.setChassisControl(goipmi.ControlPowerDown)
}

func (o *outBand) PowerReset() error {
	return o.setChassisControl(goipmi.ControlPowerHardReset)
}

func (o *outBand) PowerCycle() error {
	return o.setChassisControl(goipmi.ControlPowerCycle)
}

func (o *outBand) setChassisControl(chassisControl goipmi.ChassisControl) error {
	client, err := ipmi.OpenClientConnection(o.Connection())
	if err != nil {
		return err
	}
	defer client.Close()
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
	return o.setChassisIdentify(0x01)
}

func (o *outBand) IdentifyLEDOff() error {
	return o.setChassisIdentify(0x00)
}

func (o *outBand) setChassisIdentify(forceOn uint8) error {
	client, err := ipmi.OpenClientConnection(o.Connection())
	if err != nil {
		return err
	}
	defer client.Close()
	return ipmi.SetChassisIdentify(client, forceOn)
}

func (o *outBand) BootFrom(bootTarget hal.BootTarget) error {
	client, err := ipmi.OpenClientConnection(o.Connection())
	if err != nil {
		return err
	}
	defer client.Close()

	useProgress := true
	// set set-in-progress flag
	err = ipmi.SetSystemBoot(client, goipmi.BootParamSetInProgress, 1)
	if err != nil {
		useProgress = false
	}

	err = ipmi.SetSystemBoot(client, goipmi.BootParamInfoAck, 1, 1)
	if err != nil {
		if useProgress {
			// set-in-progress = set-complete
			_ = ipmi.SetSystemBoot(client, goipmi.BootParamSetInProgress, 0)
		}
		return err
	}

	uefiQualifier, bootDevQualifier := ipmi.GetBootOrderQualifiers(bootTarget, o.compliance)
	err = ipmi.SetSystemBoot(client, goipmi.BootParamBootFlags, uefiQualifier, bootDevQualifier, 0, 0, 0)
	if err == nil {
		if useProgress {
			// set-in-progress = commit-write
			_ = ipmi.SetSystemBoot(client, goipmi.BootParamSetInProgress, 2)
		}
	}

	if useProgress {
		// set-in-progress = set-complete
		_ = ipmi.SetSystemBoot(client, goipmi.BootParamSetInProgress, 0)
	}

	return err
}

func (o *outBand) Describe() string {
	return "OutBand connected to Supermicro"
}

func (o *outBand) Connection() (string, string, string) {
	return o.ip, o.user, o.password
}
