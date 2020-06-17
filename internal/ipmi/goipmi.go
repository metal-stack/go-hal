package ipmi

import (
	"errors"
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

func (cc *ClientConnection) SetChassisIdentify(forceOn uint8) error {
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

// ChassisControlRequest per section 28.3
type ChassisControlRequest struct {
	control uint8
}

// ChassisControlResponse per section 28.3
type ChassisControlResponse struct {
	goipmi.CompletionCode
}

func (cc *ClientConnection) SetChassisControl(control goipmi.ChassisControl) error {
	r := &goipmi.Request{
		NetworkFunction: goipmi.NetworkFunctionChassis,
		Command:         goipmi.Command(ChassisIdentify),
		Data: &ChassisControlRequest{
			control: uint8(control),
		},
	}
	resp := &ChassisControlResponse{}
	err := cc.Send(r, resp)
	if err != nil {
		return err
	}
	if goipmi.CompletionCode(resp.CompletionCode.Code()) != goipmi.CommandCompleted {
		return errors.New(resp.Error())
	}
	return nil
}
