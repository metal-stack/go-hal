package ipmi

import (
	"errors"
	goipmi "github.com/vmware/goipmi"
)

func OpenClientConnection(ip, user, password string) (*goipmi.Client, error) {
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
	return client, nil
}

func SetSystemBoot(client *goipmi.Client, param uint8, data ...uint8) error {
	r := &goipmi.Request{
		NetworkFunction: goipmi.NetworkFunctionChassis,
		Command:         goipmi.CommandSetSystemBootOptions,
		Data: &goipmi.SetSystemBootOptionsRequest{
			Param: param,
			Data:  data,
		},
	}
	return client.Send(r, &goipmi.SetSystemBootOptionsResponse{})
}

// ChassisIdentifyRequest per section 28.5
type ChassisIdentifyRequest struct {
	ForceOnOption uint8
	ForceOn       uint8
}

// ChassisIdentifyResponse per section 28.5
type ChassisIdentifyResponse struct {
	goipmi.CompletionCode
}

func SetChassisIdentify(client *goipmi.Client, forceOn uint8) error {
	r := &goipmi.Request{
		NetworkFunction: goipmi.NetworkFunctionChassis,
		Command:         goipmi.Command(ChassisIdentify),
		Data: &ChassisIdentifyRequest{
			ForceOnOption: 0,
			ForceOn:       forceOn,
		},
	}
	resp := &ChassisIdentifyResponse{}
	err := client.Send(r, resp)
	if err != nil {
		return err
	}
	if goipmi.CompletionCode(resp.CompletionCode.Code()) != goipmi.CommandCompleted {
		return errors.New(resp.Error())
	}
	return nil
}
