package redfish

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/logger"
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
	chassis, err := c.Service.Chassis()
	if err != nil {
		c.log.Warnf("ignore chassis query err:%s\n", err.Error())
	}

	for _, chass := range chassis {
		c.log.Debugf("Model:" + chass.Model + " Name:" + chass.Name + " Part:" + chass.PartNumber + " Serial:" + chass.SerialNumber + " SKU:" + chass.SKU + "\n")
		if chass.ChassisType == redfish.RackMountChassisType {
			return &api.Board{
				VendorString: chass.Manufacturer,
				Model:        chass.Model,
				PartNumber:   chass.PartNumber,
				SerialNumber: chass.SerialNumber,
			}, nil
		}
	}
	return nil, fmt.Errorf("no board detected: #chassis:%d", len(chassis))
}

// MachineUUID retrieves a unique uuid for this (hardware) machine
func (c *APIClient) MachineUUID() (string, error) {
	systems, err := c.Service.Systems()
	if err != nil {
		c.log.Errorf("ignored system query err:%s\n", err.Error())
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
		c.log.Warnf("ignored system query err:%s\n", err.Error())
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
		c.log.Warnf("ignored system query err:%s\n", err.Error())
	}
	for _, system := range systems {
		err = system.Reset(resetType)
		if err == nil {
			return nil
		}
	}
	return errors.Wrapf(err, "failed to set power to %s", resetType)
}

func (c *APIClient) SetBootOrder(target hal.BootTarget, vendor api.Vendor) error {
	if target == hal.BootTargetBIOS {
		return c.setNextBootBIOS()
	}

	currentBootOrder, err := c.retrieveBootOrder(vendor)
	if err != nil {
		return err
	}
	switch target {
	default:
		return c.setPersistentPXE(currentBootOrder)
	case hal.BootTargetDisk:
		return c.setPersistentHDD(currentBootOrder)
	}
}

func (c *APIClient) retrieveBootOrder(vendor api.Vendor) ([]string, error) { //TODO move out
	if vendor != api.VendorLenovo { // TODO implement also for Supermicro
		return nil, fmt.Errorf("retrieveBootOrder via Redfish is not yet implemented for vendor %q", vendor)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/Systems/1/Oem/Lenovo/BootSettings/BootOrder.BootOrder", c.urlPrefix), nil)
	if err != nil {
		return nil, err
	}
	c.addHeadersAndAuth(req)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	type boot struct {
		BootOrderCurrent []string `json:"BootOrderCurrent"`
	}
	b := boot{}
	err = json.Unmarshal(buf.Bytes(), &b)
	return b.BootOrderCurrent, err
}

func (c *APIClient) setPersistentPXE(bootOrder []string) error {
	sort.SliceStable(bootOrder, func(i, j int) bool {
		if strings.ToLower(bootOrder[i]) == "network" {
			return true
		}
		return !strings.Contains(bootOrder[i], "metal")
	})
	return c.setBootOrder(bootOrder)
}

func (c *APIClient) setPersistentHDD(bootOrder []string) error {
	sort.SliceStable(bootOrder, func(i, j int) bool {
		return strings.Contains(bootOrder[i], "metal")
	})
	return c.setBootOrder(bootOrder)
}

func (c *APIClient) setBootOrder(bootOrder []string) error {
	type boot struct {
		BootOrderNext []string `json:"BootOrderNext"`
	}
	body, err := json.Marshal(&boot{
		BootOrderNext: bootOrder,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/Systems/1/Oem/Lenovo/BootSettings/BootOrder.BootOrder", c.urlPrefix), bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.addHeadersAndAuth(req)
	_, err = c.Do(req)
	return err
}

func (c *APIClient) addHeadersAndAuth(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+c.basicAuth)
	req.SetBasicAuth(c.user, c.password)
}

func (c *APIClient) setNextBootBIOS() error {
	systems, err := c.Service.Systems()
	if err != nil {
		c.log.Warnf("ignored system query err:%s\n", err.Error())
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
		c.log.Warnf("ignore service query err:%s\n", err.Error())
	}

	chassis, err := c.Service.Chassis()
	if err != nil {
		c.log.Warnf("ignore chassis query err:%s\n", err.Error())
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
