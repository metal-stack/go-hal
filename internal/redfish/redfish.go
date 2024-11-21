package redfish

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/pkg/logger"

	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

type APIClient struct {
	Gofish *gofish.APIClient
	*http.Client
	urlPrefix      string
	user           string
	password       string
	basicAuth      string
	log            logger.Logger
	redfishVersion string
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
		Gofish:         c,
		Client:         c.HTTPClient,
		redfishVersion: c.Service.RedfishVersion,
		user:           user,
		password:       password,
		basicAuth:      base64.StdEncoding.EncodeToString([]byte(user + ":" + password)),
		urlPrefix:      fmt.Sprintf("%s/redfish/v1", url),
		log:            log,
	}, nil
}

func (c *APIClient) VendorAndModel() (api.Vendor, string, string, error) {
	manufacturer := ""
	model := ""
	vendor := api.VendorUnknown
	// Query the chassis data using the session token
	if c.Gofish.Service == nil {
		return vendor, manufacturer, model, fmt.Errorf("gofish service root is not available most likely due to missing username")
	}

	systems, err := c.Gofish.Service.Systems()
	if err != nil {
		c.log.Warnw("ignore system query", "error", err.Error())
	}
	for _, system := range systems {
		if system.Manufacturer != "" {
			manufacturer = system.Manufacturer
		}
		if system.Model != "" {
			model = system.Model
		}
		if manufacturer != "" && model != "" {
			break
		}
	}
	if manufacturer == "" || model == "" {
		return vendor, manufacturer, model, fmt.Errorf("unable to detect vendor and model")
	}

	vendor = api.GuessVendor(manufacturer)
	return vendor, manufacturer, model, nil
}

func (c *APIClient) BoardInfo() (*api.Board, error) {
	// Query the chassis data using the session token
	if c.Gofish.Service == nil {
		return nil, fmt.Errorf("gofish service root is not available most likely due to missing username")
	}

	biosVersion := ""
	manufacturer := ""
	model := ""

	systems, err := c.Gofish.Service.Systems()
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

	chassis, err := c.Gofish.Service.Chassis()
	if err != nil {
		c.log.Warnw("ignore system query", "error", err.Error())
	}
	for _, chass := range chassis {
		if chass.ChassisType == redfish.RackMountChassisType {
			power, err := chass.Power()
			var powerMetric *api.PowerMetric
			if err != nil {
				c.log.Warnw("ignoring power detection", "error", err)
			} else {
				for _, pc := range power.PowerControl {
					powerMetric = &api.PowerMetric{
						AverageConsumedWatts: pc.PowerMetrics.AverageConsumedWatts,
						IntervalInMin:        pc.PowerMetrics.IntervalInMin,
						MaxConsumedWatts:     pc.PowerMetrics.MaxConsumedWatts,
						MinConsumedWatts:     pc.PowerMetrics.MinConsumedWatts,
					}
					c.log.Debugw("power consumption", "metrics", powerMetric)
					break
				}
			}
			var powerSupplies []api.PowerSupply
			for _, ps := range power.PowerSupplies {
				powerSupplies = append(powerSupplies, api.PowerSupply{
					Status: api.Status{
						Health: string(ps.Status.Health),
						State:  string(ps.Status.State),
					},
				})
				c.log.Debugw("powersupplies", "powersupply", ps)
			}
			c.log.Debugw("got chassis",
				"Manufacturer", manufacturer, "Model", model, "Name", chass.Name,
				"PartNumber", chass.PartNumber, "SerialNumber", chass.SerialNumber,
				"BiosVersion", biosVersion, "led", chass.IndicatorLED)
			return &api.Board{
				VendorString:   manufacturer,
				Model:          model,
				PartNumber:     chass.PartNumber,
				SerialNumber:   chass.SerialNumber,
				BiosVersion:    biosVersion,
				RedfishVersion: c.redfishVersion,
				IndicatorLED:   toMetalLEDState(chass.IndicatorLED),
				PowerMetric:    powerMetric,
				PowerSupplies:  powerSupplies,
			}, nil
		}
	}
	return nil, fmt.Errorf("no board detected: #chassis:%d", len(chassis))
}

func toMetalLEDState(state common.IndicatorLED) string {
	switch state {
	case common.BlinkingIndicatorLED, common.LitIndicatorLED:
		return "LED-ON"
	case common.OffIndicatorLED, common.UnknownIndicatorLED:
		return "LED-OFF"
	default:
		return "LED-OFF"
	}
}

// MachineUUID retrieves a unique uuid for this (hardware) machine
func (c *APIClient) MachineUUID() (string, error) {
	systems, err := c.Gofish.Service.Systems()
	if err != nil {
		c.log.Errorw("error during system query, unable to detect uuid", "error", err.Error())
		return "", err
	}
	for _, system := range systems {
		if system.UUID != "" {
			return system.UUID, nil
		}
	}
	return "", fmt.Errorf("failed to detect machine UUID")
}

func (c *APIClient) PowerState() (hal.PowerState, error) {
	systems, err := c.Gofish.Service.Systems()
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
	state, err := c.PowerState()
	if err != nil {
		return err
	}
	if state == hal.PowerOnState {
		return nil
	}
	return c.setPower(redfish.OnResetType)
}

func (c *APIClient) PowerOff() error {
	state, err := c.PowerState()
	if err != nil {
		return err
	}
	if state == hal.PowerOffState {
		return nil
	}
	return c.setPower(redfish.ForceOffResetType)
}

func (c *APIClient) PowerReset() error {
	return c.setPower(redfish.ForceRestartResetType)
}

func (c *APIClient) PowerCycle() error {
	return c.setPower(redfish.PowerCycleResetType)
}

func (c *APIClient) setPower(resetType redfish.ResetType) error {
	systems, err := c.Gofish.Service.Systems()
	if err != nil {
		c.log.Warnw("ignore system query", "error", err.Error())
	}
	for _, system := range systems {
		err = system.Reset(resetType)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("failed to set power to %s %w", resetType, err)
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
	case hal.BootTargetDisk:
		return c.setPersistentHDD(currentBootOrder)
	case hal.BootTargetPXE, hal.BootTargetBIOS:
		fallthrough
	default:
		return c.setPersistentPXE(currentBootOrder)
	}
}

func (c *APIClient) retrieveBootOrder(vendor api.Vendor) ([]string, error) { //TODO move out
	if vendor != api.VendorLenovo { // TODO implement also for Supermicro
		return nil, fmt.Errorf("retrieveBootOrder via Redfish is not yet implemented for vendor %q", vendor)
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", fmt.Sprintf("%s/Systems/1/Oem/Lenovo/BootSettings/BootOrder.BootOrder", c.urlPrefix), nil)
	if err != nil {
		return nil, err
	}
	c.addHeadersAndAuth(req)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
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
	req, err := http.NewRequestWithContext(context.Background(), "PATCH", fmt.Sprintf("%s/Systems/1/Oem/Lenovo/BootSettings/BootOrder.BootOrder", c.urlPrefix), bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.addHeadersAndAuth(req)
	resp, err := c.Do(req)
	if err == nil {
		_ = resp.Body.Close()
	}
	return err
}

func (c *APIClient) addHeadersAndAuth(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+c.basicAuth)
	req.SetBasicAuth(c.user, c.password)
}

func (c *APIClient) setNextBootBIOS() error {
	systems, err := c.Gofish.Service.Systems()
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
	return fmt.Errorf("failed to set next boot BIOS %w", err)
}

func (c *APIClient) BMC() (*api.BMC, error) {
	systems, err := c.Gofish.Service.Systems()
	if err != nil {
		c.log.Warnw("ignore system query", "error", err.Error())
	}

	chassis, err := c.Gofish.Service.Chassis()
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

func (c *APIClient) IdentifyLEDState(state hal.IdentifyLEDState) error {
	var payload map[string]any

	switch state {
	case hal.IdentifyLEDStateOff:
		payload = map[string]any{"LocationIndicatorActive": false}
	case hal.IdentifyLEDStateOn:
		payload = map[string]any{"LocationIndicatorActive": true}
	case hal.IdentifyLEDStateBlinking:
		payload = map[string]any{"LocationIndicatorActive": true}
	case hal.IdentifyLEDStateUnknown:
		return fmt.Errorf("unknown LED state:%s", state)
	}
	_, err := c.Gofish.Patch("/redfish/v1/Chassis/System.Embedded.1", payload)
	if err != nil {
		c.log.Errorw("unable to set led", "error", err)
	}

	return nil
}

func (c *APIClient) GetIdentifyLED() (hal.IdentifyLEDState, error) {

	resp, err := c.Gofish.Get("/redfish/v1/Chassis/System.Embedded.1")
	if err != nil {
		c.log.Errorw("unable to get led", "error", err)
		return hal.IdentifyLEDStateUnknown, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return hal.IdentifyLEDStateUnknown, err
	}

	var parsedBody map[string]any
	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		c.log.Errorw("unable to parse led state", "error", err)
		return hal.IdentifyLEDStateUnknown, err
	}

	state, ok := parsedBody["LocationIndicatorActive"]
	if !ok {
		return hal.IdentifyLEDStateUnknown, fmt.Errorf("ledstate does not contain a LocationIndicatorActive key")
	}

	switch state.(bool) {
	case true:
		return hal.IdentifyLEDStateOn, nil
	case false:
		return hal.IdentifyLEDStateOff, nil
	}
	return hal.IdentifyLEDStateUnknown, fmt.Errorf("unknown state:%s", state)
}

func (c *APIClient) IdentifyLEDOn() error {
	return c.IdentifyLEDState(hal.IdentifyLEDStateOn)
}

func (c *APIClient) IdentifyLEDOff() error {
	return c.IdentifyLEDState(hal.IdentifyLEDStateOff)
}
