package redfish

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/pkg/logger"
	"github.com/pkg/errors"

	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

type APIClient struct {
	*gofish.APIClient
	*http.Client
	urlPrefix string
	user      string
	password  string
	basicAuth string
	log       logger.Logger
}

func New(url, user, password string, insecure bool, log logger.Logger) (*APIClient, error) {
	// Create a new instance of gofish and redfish client, ignoring self-signed certs
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
		APIClient: c,
		Client:    c.HTTPClient,
		user:      user,
		password:  password,
		basicAuth: base64.StdEncoding.EncodeToString([]byte(user + ":" + password)),
		urlPrefix: fmt.Sprintf("%s/redfish/v1", url),
		log:       log,
	}, nil
}

func (c *APIClient) BoardInfo() (*api.Board, error) {
	// Query the chassis data using the session token
	if c.Service == nil {
		return nil, errors.New("gofish service root is not available most likely due to missing username")
	}

	biosVersion := ""
	manufacturer := ""
	model := ""
	systems, err := c.Service.Systems()
	if err != nil {
		c.log.Warnw("ignore system query", "error", err.Error())
	}
	for _, system := range systems {
		if system.BIOSVersion != "" {
			biosVersion = system.BIOSVersion
			break
		}
	}
	for _, system := range systems {
		if system.Manufacturer != "" {
			manufacturer = system.Manufacturer
			break
		}
	}
	for _, system := range systems {
		if system.Model != "" {
			model = system.Model
			break
		}
	}

	chassis, err := c.Service.Chassis()
	if err != nil {
		c.log.Warnw("ignore system query", "error", err.Error())
	}
	for _, chass := range chassis {
		if chass.ChassisType == redfish.RackMountChassisType {
			c.log.Debugw("got chassis",
				"Manufacturer", manufacturer, "Model", model, "Name", chass.Name,
				"PartNumber", chass.PartNumber, "SerialNumber", chass.SerialNumber, "BiosVersion", biosVersion)
			return &api.Board{
				VendorString: manufacturer,
				Model:        model,
				PartNumber:   chass.PartNumber,
				SerialNumber: chass.SerialNumber,
				BiosVersion:  biosVersion,
			}, nil
		}
	}
	return nil, fmt.Errorf("no board detected: #chassis:%d", len(chassis))
}

// MachineUUID retrieves a unique uuid for this (hardware) machine
func (c *APIClient) MachineUUID() (string, error) {
	systems, err := c.Service.Systems()
	if err != nil {
		c.log.Errorw("error during system query, unable to detect uuid", "error", err.Error())
		return "", err
	}
	for _, system := range systems {
		if system.UUID != "" {
			return system.UUID, nil
		}
	}
	return "", errors.New("failed to detect machine UUID")
}

func (c *APIClient) PowerState() (hal.PowerState, error) {
	systems, err := c.Service.Systems()
	if err != nil {
		c.log.Warnw("ignore system query", "error", err.Error())
	}
	for _, system := range systems {
		if system.PowerState != "" {
			return hal.GuessPowerState(string(system.PowerState)), nil
		}
	}
	return hal.PowerUnknownState, nil
}

func (c *APIClient) PowerOn() error {
	return c.setPower(redfish.ForceOnResetType)
}

func (c *APIClient) PowerOff() error {
	return c.setPower(redfish.ForceOffResetType)
}

func (c *APIClient) PowerReset() error {
	return c.setPower(redfish.ForceRestartResetType)
}

func (c *APIClient) PowerCycle() error {
	return c.setPower(redfish.PowerCycleResetType)
}

func (c *APIClient) setPower(resetType redfish.ResetType) error {
	systems, err := c.Service.Systems()
	if err != nil {
		c.log.Warnw("ignore system query", "error", err.Error())
	}
	for _, system := range systems {
		err = system.Reset(resetType)
		if err == nil {
			return nil
		}
	}
	return errors.Wrapf(err, "failed to set power to %s", resetType)
}

func (c *APIClient) DoGet(path string) (*http.Response, error) {
	return c.do(http.MethodGet, path, nil)
}

func (c *APIClient) DoPatch(path string, body io.Reader) (*http.Response, error) {
	return c.do(http.MethodPatch, path, body)
}

func (c *APIClient) do(method, path string, body io.Reader) (*http.Response, error) {
	path = strings.TrimPrefix(path, "/")
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", c.urlPrefix, path), body)
	if err != nil {
		return nil, err
	}
	c.addHeadersAndAuth(req)
	return c.Do(req)
}

func (c *APIClient) SetBootOrder(bootOrder []string) error {
	type boot struct {
		BootOrderNext []string `json:"BootOrderNext"`
	}
	body, err := json.Marshal(&boot{
		BootOrderNext: bootOrder,
	})
	if err != nil {
		return err
	}
	_, err = c.DoPatch("/Systems/1/Oem/Lenovo/BootSettings/BootOrder.BootOrder", bytes.NewReader(body))
	return err
}

func (c *APIClient) addHeadersAndAuth(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+c.basicAuth)
	req.SetBasicAuth(c.user, c.password)
}

func (c *APIClient) SetNextBootBIOS() error {
	systems, err := c.Service.Systems()
	if err != nil {
		c.log.Warnw("ignore system query", "error", err.Error())
	}
	for _, system := range systems {
		boot := system.Boot
		boot.BootSourceOverrideTarget = redfish.BiosSetupBootSourceOverrideTarget
		boot.BootSourceOverrideEnabled = redfish.OnceBootSourceOverrideEnabled
		err = system.SetBoot(boot)
		if err == nil {
			return nil
		}
	}
	return errors.Wrap(err, "failed to set next boot BIOS")
}

func (c *APIClient) BMC() (*api.BMC, error) {
	systems, err := c.Service.Systems()
	if err != nil {
		c.log.Warnw("ignore system query", "error", err.Error())
	}

	chassis, err := c.Service.Chassis()
	if err != nil {
		c.log.Warnw("ignore system query", "error", err.Error())
	}

	bmc := &api.BMC{}

	for _, system := range systems {
		bmc.ProductManufacturer = system.Manufacturer
		bmc.ProductPartNumber = system.PartNumber
		bmc.ProductSerial = system.SerialNumber
	}

	for _, chass := range chassis {
		if chass.ChassisType != redfish.RackMountChassisType {
			continue
		}

		bmc.ChassisPartNumber = chass.PartNumber
		bmc.ChassisPartSerial = chass.SerialNumber

		bmc.BoardMfg = chass.Manufacturer
	}

	//TODO find bmc.BoardMfgSerial and bmc.BoardPartNumber

	return bmc, nil
}
