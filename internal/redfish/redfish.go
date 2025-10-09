package redfish

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/pkg/logger"

	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/common"
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

type bootOverrideRequest struct {
	Boot redfish.Boot `json:"Boot"`
}

type bootOrderSetRequest struct {
	Boot struct {
		BootOrder []string `json:"BootOrder"`
	} `json:"Boot"`
}

type indicatorLEDRequest struct {
	IndicatorLED common.IndicatorLED `json:"IndicatorLED"`
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
		return nil, fmt.Errorf("gofish service root is not available most likely due to missing username")
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
				VendorString:  manufacturer,
				Model:         model,
				PartNumber:    chass.PartNumber,
				SerialNumber:  chass.SerialNumber,
				BiosVersion:   biosVersion,
				IndicatorLED:  toMetalLEDState(chass.IndicatorLED),
				PowerMetric:   powerMetric,
				PowerSupplies: powerSupplies,
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
	return "", fmt.Errorf("failed to detect machine UUID")
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
	return fmt.Errorf("failed to set power to %s %w", resetType, err)
}

// SetChassisIdentifyLEDState sets the chassis identify LED to given state
func (c *APIClient) SetChassisIdentifyLEDState(state hal.IdentifyLEDState) error {
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

// SetChassisIdentifyLEDOn turns on the chassis identify LED
func (c *APIClient) SetChassisIdentifyLEDOn() error {
	payload := indicatorLEDRequest{
		IndicatorLED: common.LitIndicatorLED,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPatch, fmt.Sprintf("%s/Chassis/1", c.urlPrefix), bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.addHeadersAndAuth(req)

	resp, err := c.Do(req)
	if err == nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("unable to turn on the chassis identify LED %w", err)
	}
	return nil
}

// SetChassisIdentifyLEDOff turns off the chassis identify LED
func (c *APIClient) SetChassisIdentifyLEDOff() error {
	payload := indicatorLEDRequest{
		IndicatorLED: common.OffIndicatorLED,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPatch, fmt.Sprintf("%s/Chassis/1", c.urlPrefix), bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.addHeadersAndAuth(req)

	resp, err := c.Do(req)
	if err == nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("unable to turn off the chassis identify LED %w", err)
	}
	return nil
}

func (c *APIClient) SetBootTarget(target hal.BootTarget) error {
	switch target {
	case hal.BootTargetBIOS:
		return c.setNextBootBIOS()
	case hal.BootTargetDisk:
		return c.setPersistentHDD()
	case hal.BootTargetPXE:
		fallthrough
	default:
		return c.setPersistentPXE()
	}
}

func (c *APIClient) setPersistentPXE() error {
	payload := bootOverrideRequest{
		Boot: redfish.Boot{
			BootSourceOverrideEnabled: redfish.ContinuousBootSourceOverrideEnabled,
			BootSourceOverrideMode:    redfish.UEFIBootSourceOverrideMode,
			BootSourceOverrideTarget:  redfish.PxeBootSourceOverrideTarget,
		},
	}
	return c.setBootTargetOverride(payload)
}

func (c *APIClient) setPersistentHDD() error {
	payload := bootOverrideRequest{
		Boot: redfish.Boot{
			BootSourceOverrideEnabled: redfish.ContinuousBootSourceOverrideEnabled,
			BootSourceOverrideMode:    redfish.UEFIBootSourceOverrideMode,
			BootSourceOverrideTarget:  redfish.HddBootSourceOverrideTarget,
		},
	}
	return c.setBootTargetOverride(payload)
}

func (c *APIClient) setBootTargetOverride(payload bootOverrideRequest) error {
	systems, err := c.Service.Systems()
	if err != nil {
		return fmt.Errorf("unable to query systems: %w", err)
	}

	if len(systems) == 0 {
		return fmt.Errorf("no system found to set boot target")
	}

	if len(systems) > 1 {
		c.log.Warnw("multiple systems found, ignoring all but the first one", "count", len(systems))
	}

	// Assuming there's typically one primary system.
	system := systems[0]

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPatch, fmt.Sprintf("%s/Systems/%s", c.urlPrefix, system.ID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.addHeadersAndAuth(req)

	resp, err := c.Do(req)
	_ = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("unable to override boot order %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to override boot order, http status: %s", resp.Status)
	}
	return nil
}

func (c *APIClient) addHeadersAndAuth(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+c.basicAuth)
	req.Header.Add("If-Match", "*")
	req.SetBasicAuth(c.user, c.password)
}

func (c *APIClient) setNextBootBIOS() error {
	payload := bootOverrideRequest{
		Boot: redfish.Boot{
			BootSourceOverrideEnabled: redfish.OnceBootSourceOverrideEnabled,
			BootSourceOverrideMode:    redfish.UEFIBootSourceOverrideMode,
			BootSourceOverrideTarget:  redfish.BiosSetupBootSourceOverrideTarget,
		},
	}
	return c.setBootTargetOverride(payload)
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

func (c *APIClient) GetBootOptions() ([]*redfish.BootOption, error) {
	// The curl command here would be curl -k -u <user>:<pwd> https://10.1.1.18/redfish/v1/Systems/System.Embedded.1/BootOptions
	systems, err := c.Service.Systems()
	if err != nil {
		c.log.Warnw("ignore system query", "error", err.Error())
	}
	for _, system := range systems {
		bootOptions, err := system.BootOptions()
		if err != nil {
			c.log.Warnw("ignore boot options query", "error", err.Error())
			continue
		}
		if len(bootOptions) == 0 {
			c.log.Warnw("no boot options found", "error")
			continue
		}
		if len(system.Boot.BootOrder) == 0 {
			c.log.Warnw("no boot order found", "error")
			continue
		}
		return bootOptions, nil
	}

	return nil, fmt.Errorf("failed to get boot options")
}

// SetBootOrder sets the boot order to match the sequence of the boot option entries
func (c *APIClient) SetBootOrder(entries []*redfish.BootOption) error {
	systems, err := c.Service.Systems()
	if err != nil {
		return fmt.Errorf("unable to query systems: %w", err)
	}

	if len(systems) == 0 {
		return fmt.Errorf("no system found to set boot order")
	}

	if len(systems) > 1 {
		c.log.Warnw("multiple systems found, ignoring all but the first one", "count", len(systems))
	}

	// Assuming there's typically one primary system.
	system := systems[0]

	var bootOrder []string
	for _, entry := range entries {
		bootOrder = append(bootOrder, entry.ID)
	}

	if len(system.Boot.BootOrder) == 0 {
		c.log.Errorw("no boot order found")
	}

	payload := bootOrderSetRequest{}
	payload.Boot.BootOrder = bootOrder

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPatch, fmt.Sprintf("%s/Systems/%s", c.urlPrefix, system.ID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.addHeadersAndAuth(req)
	resp, err := c.Do(req)
	_ = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("unable to set boot order: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to set boot order, http status: %s", resp.Status)
	}

	return nil
}
