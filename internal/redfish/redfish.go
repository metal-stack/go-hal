package redfish

import (
	"fmt"

	"github.com/metal-stack/go-hal/internal/api"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

type APIClient struct {
	c *gofish.APIClient
}

func New(url, user, password string, insecure bool) (*APIClient, error) {
	// Create a new instance of gofish client, ignoring self-signed certs
	config := gofish.ClientConfig{
		Endpoint: url,
		Username: user,
		Password: password,
		Insecure: insecure,
	}
	c, err := gofish.Connect(config)
	if err != nil {
		return nil, err
	}
	return &APIClient{
		c: c,
	}, nil
}

func (c *APIClient) BoardInfo() (*api.Board, error) {
	// Retrieve the service root
	service := c.c.Service

	// Query the chassis data using the session token
	chassis, err := service.Chassis()
	if err != nil {
		return nil, err
	}

	for _, chass := range chassis {
		if chass.ChassisType == redfish.RackMountChassisType {
			return &api.Board{
				Vendor: chass.Manufacturer,
				Name:   chass.Model,
			}, nil
		}
	}
	return nil, fmt.Errorf("no board detected")
}

// MachineUUID retrieves a unique uuid for this (hardware) machine
func (c *APIClient) MachineUUID() (string, error) {
	service := c.c.Service

	systems, err := service.Systems()
	if err != nil {
		return "00000000-0000-0000-0000-000000000000", err
	}
	for _, system := range systems {
		if system.UUID != "" {
			return system.UUID, nil
		}
	}
	return "00000000-0000-0000-0000-000000000000", err
}
