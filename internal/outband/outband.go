package outband

import (
	"fmt"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/pkg/api"
	goipmi "github.com/vmware/goipmi"
)

type OutBand struct {
	Redfish    *redfish.APIClient
	board      *api.Board
	compliance api.Compliance
	ip         string
	user       string
	password   string
}

func New(r *redfish.APIClient, board *api.Board, compliance api.Compliance, ip, user, password string) *OutBand {
	return &OutBand{
		Redfish:    r,
		board:      board,
		compliance: compliance,
		ip:         ip,
		user:       user,
		password:   password,
	}
}

func (ob *OutBand) Board() *api.Board {
	return ob.board
}

func (ob *OutBand) Connection() (string, string, string) {
	return ob.ip, ob.user, ob.password
}

func (ob *OutBand) SetChassisControl(control goipmi.ChassisControl) error {
	client, err := ipmi.OpenClientConnection(ob.Connection())
	if err != nil {
		return err
	}
	defer client.Close()
	return client.SetChassisControl(control)
}

func (ob *OutBand) SetChassisIdentifyLEDState(state hal.IdentifyLEDState) error {
	switch state {
	case hal.IdentifyLEDStateOff:
		return ob.SetChassisIdentifyLEDOff()
	case hal.IdentifyLEDStateOn:
		return ob.SetChassisIdentifyLEDOn()
	default:
		return fmt.Errorf("unknown identify LED state: %s", state)
	}
}

func (ob *OutBand) SetChassisIdentifyLEDOff() error {
	return ob.SetChassisIdentify(ipmi.False)
}

func (ob *OutBand) SetChassisIdentifyLEDOn() error {
	return ob.SetChassisIdentify(ipmi.True)
}

func (ob *OutBand) SetChassisIdentify(forceOn uint8) error {
	client, err := ipmi.OpenClientConnection(ob.Connection())
	if err != nil {
		return err
	}
	defer client.Close()
	return client.SetChassisIdentify(forceOn)
}

func (ob *OutBand) SetBootOrder(bootTarget hal.BootTarget) error {
	client, err := ipmi.OpenClientConnection(ob.Connection())
	if err != nil {
		return err
	}
	defer client.Close()

	useProgress := true
	// set set-in-progress flag
	err = client.SetSystemBoot(goipmi.BootParamSetInProgress, 1)
	if err != nil {
		useProgress = false
	}

	err = client.SetSystemBoot(goipmi.BootParamInfoAck, 1, 1)
	if err != nil {
		if useProgress {
			// set-in-progress = set-complete
			_ = client.SetSystemBoot(goipmi.BootParamSetInProgress, 0)
		}
		return err
	}

	uefiQualifier, bootDevQualifier := ipmi.GetBootOrderQualifiers(bootTarget, ob.compliance)
	err = client.SetSystemBoot(goipmi.BootParamBootFlags, uefiQualifier, bootDevQualifier, 0, 0, 0)
	if err == nil {
		if useProgress {
			// set-in-progress = commit-write
			_ = client.SetSystemBoot(goipmi.BootParamSetInProgress, 2)
		}
	}

	if useProgress {
		// set-in-progress = set-complete
		_ = client.SetSystemBoot(goipmi.BootParamSetInProgress, 0)
	}

	return err
}
