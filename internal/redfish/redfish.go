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
func (c *APIClient) SetChassisIdentifyLEDState(state hal.IdentifyLEDState, vendor api.Vendor) error {
	if vendor != api.VendorGigabyte {
		return fmt.Errorf("setChassisIdentifyLEDOn via Redfish is not yet implemented for vendor %q", vendor)
	}

	switch state {
	case hal.IdentifyLEDStateOff:
		return c.SetChassisIdentifyLEDOff(vendor)
	case hal.IdentifyLEDStateOn:
		return c.SetChassisIdentifyLEDOn(vendor)
	case hal.IdentifyLEDStateUnknown:
		fallthrough
	default:
		return fmt.Errorf("unknown identify LED state: %s", state)
	}
}

// SetChassisIdentifyLEDOn turns on the chassis identify LED
func (c *APIClient) SetChassisIdentifyLEDOn(vendor api.Vendor) error {
	if vendor != api.VendorGigabyte {
		return fmt.Errorf("setChassisIdentifyLEDOn via Redfish is not yet implemented for vendor %q", vendor)
	}

	type indicatorLED struct {
		IndicatorLED string `json:"IndicatorLED"`
	}
	body, err := json.Marshal(&indicatorLED{
		IndicatorLED: "Lit",
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(context.Background(), "PATCH", fmt.Sprintf("%s/Chassis/1", c.urlPrefix), bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.addHeadersAndAuth(req)

	//ETag is the current status of the resource to be modified, ensuring that the correct resource is modified
	err = c.addEtagHeader(req)
	if err != nil {
		return err
	}
	// Alternatively, disable etag match
	//c.Service.DisableEtagMatch(true)

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
func (c *APIClient) SetChassisIdentifyLEDOff(vendor api.Vendor) error {
	if vendor != api.VendorGigabyte {
		return fmt.Errorf("setChassisIdentifyLEDOff via Redfish is not yet implemented for vendor %q", vendor)
	}

	type indicatorLED struct {
		IndicatorLED string `json:"IndicatorLED"`
	}
	body, err := json.Marshal(&indicatorLED{
		IndicatorLED: "Off",
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(context.Background(), "PATCH", fmt.Sprintf("%s/Chassis/1", c.urlPrefix), bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.addHeadersAndAuth(req)

	err = c.addEtagHeader(req)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err == nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("unable to turn off the chassis identify LED %w", err)
	}
	return nil
}

func (c *APIClient) getEtag() (string, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", fmt.Sprintf("%s/Chassis", c.urlPrefix), nil)
	if err != nil {
		return "", err
	}
	c.addHeadersAndAuth(req)
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}
	type etag struct {
		Etag string `json:"@odata.etag"`
	}
	e := etag{}
	err = json.Unmarshal(buf.Bytes(), &e)
	return e.Etag, err
}

func (c *APIClient) SetBootOrder(target hal.BootTarget, vendor api.Vendor) error {
	c.log.Infow("set boot order", "target", target, "vendor", vendor, "error", nil)
	if target == hal.BootTargetBIOS { //TODO: Implement for vendor Gigabyte
		return c.setNextBootBIOS()
	}

	switch target {
	case hal.BootTargetDisk:
		return c.setPersistentHDD(vendor)
	case hal.BootTargetPXE, hal.BootTargetBIOS:
		fallthrough
	default:
		return c.setPersistentPXE(vendor)
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

func (c *APIClient) setPersistentPXE(vendor api.Vendor) error {
	switch vendor {
	case api.VendorLenovo:
		currentBootOrder, err := c.retrieveBootOrder(vendor)
		if err != nil {
			return err
		}

		sort.SliceStable(currentBootOrder, func(i, j int) bool {
			if strings.ToLower(currentBootOrder[i]) == "network" {
				return true
			}
			return !strings.Contains(currentBootOrder[i], "metal")
		})
		return c.setBootOrder(currentBootOrder)
	case api.VendorGigabyte:
		c.log.Infow("set pxe boot order", "vendor", vendor, "error", nil)
		return c.setBootOrderOverride()
	case api.VendorUnknown, api.VendorSupermicro, api.VendorNovarion, api.VendorDell, api.VendorVagrant:
		fallthrough
	default:
		return fmt.Errorf("setPersistentPXE via Redfish is not yet implemented for vendor %q", vendor)
	}
}

func (c *APIClient) setPersistentHDD(vendor api.Vendor) error {
	switch vendor {
	case api.VendorLenovo:
		currentBootOrder, err := c.retrieveBootOrder(vendor)
		if err != nil {
			return err
		}

		sort.SliceStable(currentBootOrder, func(i, j int) bool {
			return strings.Contains(currentBootOrder[i], "metal")
		})
		return c.setBootOrder(currentBootOrder)
	case api.VendorGigabyte:
		return c.setBootOrderOverride()
	case api.VendorUnknown, api.VendorSupermicro, api.VendorNovarion, api.VendorDell, api.VendorVagrant:
		fallthrough
	default:
		return fmt.Errorf("setPersistentHDD via Redfish is not yet implemented for vendor %q", vendor)
	}
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

func (c *APIClient) setBootOrderOverride() error {
	type bootSettings struct {
		BootSourceOverrideEnabled string `json:"BootSourceOverrideEnabled,omitempty"`
		BootSourceOverrideMode    string `json:"BootSourceOverrideMode,omitempty"`
		BootSourceOverrideTarget  string `json:"BootSourceOverrideTarget,omitempty"`
	}

	type bootOverrideRequest struct {
		Boot bootSettings `json:"Boot"`
	}

	payload := bootOverrideRequest{
		Boot: bootSettings{
			BootSourceOverrideEnabled: "Disabled",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	c.log.Infow("successfully marshal", "payload", string(body))
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPatch, fmt.Sprintf("%s/Systems/Self", c.urlPrefix), bytes.NewReader(body))
	if err != nil {
		c.log.Errorw("unable to create request", "error", err.Error())
		return err
	}
	c.addHeadersAndAuth(req)

	err = c.addEtagHeader(req)
	if err != nil {
		c.log.Errorw("unable to add etag header", "error", err.Error())
		return err
	}
	c.log.Debugw("debugging request", "Header", req.Header.Get("If-Match"), "URL", req.URL.String(), "Host", req.Host)
	c.log.Infow("successfully added headers, performing request to", "host", req.Host)

	resp, err := c.Do(req)
	if err != nil {
		c.log.Errorw("error while performing request", "error", err.Error())
		return fmt.Errorf("error while performing request %w", err)
	}
	defer resp.Body.Close()

	c.log.Infow("successfully performed request, reading response", "status", resp.StatusCode)

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Errorw("error reading response body", "error", err.Error())
	}
	c.log.Infow("successfully read response body", "response", string(response))
	return nil
}

func (c *APIClient) addHeadersAndAuth(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+c.basicAuth)
	req.SetBasicAuth(c.user, c.password)
}

func (c *APIClient) addEtagHeader(req *http.Request) error {
	etag, err := c.getEtag()
	if err != nil {
		return fmt.Errorf("unable to retrieve etag %w", err)
	}
	req.Header.Add("If-Match", etag)
	return nil
}

func (c *APIClient) setNextBootBIOS() error {
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
	return fmt.Errorf("failed to set next boot BIOS %w", err)
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
