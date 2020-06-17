package ipmi

import (
	"errors"
	"fmt"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/pkg/api"
	goipmi "github.com/vmware/goipmi"
)

type ClientConnection struct {
	*goipmi.Client
}

func OpenClientConnection(ip, user, password string) (*ClientConnection, error) {
	conn := &goipmi.Connection{
		Hostname:  ip,
		Port:      623,
		Username:  user,
		Password:  password,
		Interface: "lanplus",
	}

	client, err := goipmi.NewClient(conn)
	if err != nil {
		return nil, err
	}

	err = client.Open()
	if err != nil {
		return nil, err
	}
	return &ClientConnection{
		Client: client,
	}, nil
}

func (cc *ClientConnection) SetSystemBoot(param uint8, data ...uint8) error {
	r := &goipmi.Request{
		NetworkFunction: goipmi.NetworkFunctionChassis,
		Command:         goipmi.CommandSetSystemBootOptions,
		Data: &goipmi.SetSystemBootOptionsRequest{
			Param: param,
			Data:  data,
		},
	}
	return cc.Send(r, &goipmi.SetSystemBootOptionsResponse{})
}

// ChassisIdentifyRequest per section 28.5
type ChassisIdentifyRequest struct {
	IntervalSeconds uint8
	ForceOn         uint8
}

// ChassisIdentifyResponse per section 28.5
type ChassisIdentifyResponse struct {
	goipmi.CompletionCode
}

func (cc *ClientConnection) SetChassisIdentifyLEDState(state hal.IdentifyLEDState) error {
	switch state {
	case hal.IdentifyLEDStateOff:
		return cc.SetChassisIdentifyLEDOff()
	case hal.IdentifyLEDStateOn:
		return cc.SetChassisIdentifyLEDOn()
	default:
		return fmt.Errorf("unknown identify LED state: %s", state)
	}
}

func (cc *ClientConnection) SetChassisIdentifyLEDOff() error {
	return cc.setChassisIdentifyLED(False)
}

func (cc *ClientConnection) SetChassisIdentifyLEDOn() error {
	return cc.setChassisIdentifyLED(True)
}

func (cc *ClientConnection) setChassisIdentifyLED(forceOn uint8) error {
	r := &goipmi.Request{
		NetworkFunction: goipmi.NetworkFunctionChassis,
		Command:         goipmi.Command(ChassisIdentify),
		Data: &ChassisIdentifyRequest{
			IntervalSeconds: 0,
			ForceOn:         forceOn,
		},
	}
	resp := &ChassisIdentifyResponse{}
	err := cc.Send(r, resp)
	if err != nil {
		return err
	}
	if goipmi.CompletionCode(resp.CompletionCode.Code()) != goipmi.CommandCompleted {
		return errors.New(resp.Error())
	}
	return nil
}

func (cc *ClientConnection) SetBootOrder(bootTarget hal.BootTarget, compliance api.Compliance) error {
	useProgress := true
	// set set-in-progress flag
	err := cc.SetSystemBoot(goipmi.BootParamSetInProgress, 1)
	if err != nil {
		useProgress = false
	}

	err = cc.SetSystemBoot(goipmi.BootParamInfoAck, 1, 1)
	if err != nil {
		if useProgress {
			// set-in-progress = set-complete
			_ = cc.SetSystemBoot(goipmi.BootParamSetInProgress, 0)
		}
		return err
	}

	uefiQualifier, bootDevQualifier := GetBootOrderQualifiers(bootTarget, compliance)
	err = cc.SetSystemBoot(goipmi.BootParamBootFlags, uefiQualifier, bootDevQualifier, 0, 0, 0)
	if err == nil {
		if useProgress {
			// set-in-progress = commit-write
			_ = cc.SetSystemBoot(goipmi.BootParamSetInProgress, 2)
		}
	}

	if useProgress {
		// set-in-progress = set-complete
		_ = cc.SetSystemBoot(goipmi.BootParamSetInProgress, 0)
	}

	return err
}
