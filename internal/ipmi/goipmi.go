package ipmi

import (
	"errors"
	"fmt"

	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/pkg/api"
	goipmi "github.com/vmware/goipmi"
)

type Client struct {
	*goipmi.Client
	api.Vendor
}

func OpenClientConnection(ip string, port int, user, password string) (*Client, error) {
	conn := &goipmi.Connection{
		Hostname:  ip,
		Port:      port,
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
	return &Client{
		Client: client,
	}, nil
}

func (c *Client) SetSystemBoot(param uint8, data ...uint8) error {
	r := &goipmi.Request{
		NetworkFunction: goipmi.NetworkFunctionChassis,
		Command:         goipmi.CommandSetSystemBootOptions,
		Data: &goipmi.SetSystemBootOptionsRequest{
			Param: param,
			Data:  data,
		},
	}
	return c.Send(r, &goipmi.SetSystemBootOptionsResponse{})
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

func (c *Client) SetChassisIdentifyLEDState(state hal.IdentifyLEDState) error {
	switch state {
	case hal.IdentifyLEDStateOff:
		return c.SetChassisIdentifyLEDOff()
	case hal.IdentifyLEDStateOn:
		return c.SetChassisIdentifyLEDOn()
	case hal.IdentifyLEDStateUnknown:
		fallthrough
	default:
		return fmt.Errorf("unknown identify LED state: %s", state)
	}
}

func (c *Client) SetChassisIdentifyLEDOff() error {
	return c.setChassisIdentifyLED(False)
}

func (c *Client) SetChassisIdentifyLEDOn() error {
	return c.setChassisIdentifyLED(True)
}

func (c *Client) setChassisIdentifyLED(forceOn uint8) error {
	r := &goipmi.Request{
		NetworkFunction: goipmi.NetworkFunctionChassis,
		Command:         goipmi.Command(ChassisIdentify),
		Data: &ChassisIdentifyRequest{
			IntervalSeconds: 0,
			ForceOn:         forceOn,
		},
	}
	resp := &ChassisIdentifyResponse{}
	err := c.Send(r, resp)
	if err != nil {
		return err
	}
	if goipmi.CompletionCode(resp.CompletionCode.Code()) != goipmi.CommandCompleted {
		return errors.New(resp.Error())
	}
	return nil
}

func (c *Client) SetBootOrder(bootTarget hal.BootTarget, vendor api.Vendor) error {
	useProgress := true
	// set set-in-progress flag
	err := c.SetSystemBoot(goipmi.BootParamSetInProgress, 1)
	if err != nil {
		useProgress = false
	}

	err = c.SetSystemBoot(goipmi.BootParamInfoAck, 1, 1)
	if err != nil {
		if useProgress {
			// set-in-progress = set-complete
			_ = c.SetSystemBoot(goipmi.BootParamSetInProgress, 0)
		}
		return err
	}

	uefiQualifier, bootDevQualifier := getBootOrderQualifiers(bootTarget, vendor)
	err = c.SetSystemBoot(goipmi.BootParamBootFlags, uefiQualifier, bootDevQualifier, 0, 0, 0)
	if err == nil {
		if useProgress {
			// set-in-progress = commit-write
			_ = c.SetSystemBoot(goipmi.BootParamSetInProgress, 2)
		}
	}

	if useProgress {
		// set-in-progress = set-complete
		_ = c.SetSystemBoot(goipmi.BootParamSetInProgress, 0)
	}

	return err
}
