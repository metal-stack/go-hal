package redfish

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/pkg/logger"
	"github.com/metal-stack/metal-lib/pkg/pointer"

	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/schemas"
)

type APIClient struct {
	client *gofish.APIClient
	*http.Client
	urlPrefix         string
	user              string
	password          string
	log               logger.Logger
	connectionTimeout time.Duration
	ETagRequired      bool
}

type bootOverrideRequest struct {
	Boot schemas.Boot `json:"Boot"`
}

type bootOrderSetRequest struct {
	Boot struct {
		BootOrder []string `json:"BootOrder"`
	} `json:"Boot"`
}

type indicatorLEDRequest struct {
	IndicatorLED schemas.IndicatorLED `json:"IndicatorLED"`
}

func New(url, user, password string, insecure bool, log logger.Logger, connectionTimeout *time.Duration) (*APIClient, error) {
	// Create a new instance of gofish and redfish client, ignoring self-signed certs
	config := gofish.ClientConfig{
		Endpoint: url,
		Username: user,
		Password: password,
		Insecure: insecure,
	}

	timeout := 10 * time.Second
	if connectionTimeout != nil {
		timeout = *connectionTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c, err := gofish.ConnectContext(ctx, config)
	if err != nil {
		return nil, err
	}

	return &APIClient{
		client:            c,
		Client:            c.HTTPClient,
		user:              user,
		password:          password,
		urlPrefix:         fmt.Sprintf("%s/redfish/v1", url),
		log:               log,
		connectionTimeout: timeout,
		ETagRequired:      false,
	}, nil
}

func (c *APIClient) SetETagRequired(required bool) {
	c.ETagRequired = required
}

// GetSystem returns the single managed ComputerSystem.
// Servers are expected to expose exactly one system via Redfish.
func (c *APIClient) GetSystem() (*schemas.ComputerSystem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.connectionTimeout)
	defer cancel()
	return c.getSystem(ctx)
}

// getSystem is the internal implementation of GetSystem, accepting an existing context.
func (c *APIClient) getSystem(ctx context.Context) (*schemas.ComputerSystem, error) {
	g := c.client.WithContext(ctx)
	if g.Service == nil {
		return nil, fmt.Errorf("gofish service root is not available")
	}
	systems, err := g.Service.Systems()
	if err != nil {
		return nil, fmt.Errorf("failed to query systems: %w", err)
	}
	if len(systems) == 0 {
		return nil, fmt.Errorf("no system found")
	}
	if len(systems) > 1 {
		c.log.Warnw("multiple systems found, using first one", "count", len(systems))
	}
	return systems[0], nil
}

// getChassis returns the primary chassis, preferring RackMount type.
// Servers are expected to expose exactly one chassis via Redfish.
func (c *APIClient) getChassis(ctx context.Context) (*schemas.Chassis, error) {
	g := c.client.WithContext(ctx)
	if g.Service == nil {
		return nil, fmt.Errorf("gofish service root is not available")
	}
	chassis, err := g.Service.Chassis()
	if err != nil {
		return nil, fmt.Errorf("failed to query chassis: %w", err)
	}
	for _, chass := range chassis {
		if chass.ChassisType == schemas.RackMountChassisType {
			return chass, nil
		}
	}
	if len(chassis) > 0 {
		return chassis[0], nil
	}
	return nil, fmt.Errorf("no chassis found")
}

func (c *APIClient) BoardInfo() (*api.Board, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.connectionTimeout)
	defer cancel()

	system, err := c.getSystem(ctx)
	if err != nil {
		c.log.Warnw("ignore system query", "error", err)
	}

	biosVersion := ""
	manufacturer := ""
	model := ""
	if system != nil {
		biosVersion = system.BiosVersion
		manufacturer = system.Manufacturer
		model = system.Model
	}

	chass, err := c.getChassis(ctx)
	if err != nil {
		return nil, fmt.Errorf("no board detected: %w", err)
	}

	power, err := chass.Power()
	var powerMetric *api.PowerMetric
	var powerSupplies []api.PowerSupply
	if err != nil {
		c.log.Warnw("ignoring power detection", "error", err)
	} else {
		for _, pc := range power.PowerControl {
			pm := pc.PowerMetrics
			if pm.AverageConsumedWatts == nil && pm.IntervalInMin == nil {
				continue
			}
			powerMetric = &api.PowerMetric{
				AverageConsumedWatts: pointer.SafeDeref(pm.AverageConsumedWatts),
				IntervalInMin:        float32(pointer.SafeDeref(pm.IntervalInMin)),
				MaxConsumedWatts:     pointer.SafeDeref(pm.MaxConsumedWatts),
				MinConsumedWatts:     pointer.SafeDeref(pm.MinConsumedWatts),
			}
			c.log.Debugw("power consumption", "metrics", powerMetric)
			break
		}
		for _, ps := range power.PowerSupplies {
			powerSupplies = append(powerSupplies, api.PowerSupply{
				Status: api.Status{
					Health: string(ps.Status.Health),
					State:  string(ps.Status.State),
				},
			})
			c.log.Debugw("powersupplies", "powersupply", ps)
		}
	}

	c.log.Debugw("got chassis",
		"Manufacturer", manufacturer, "Model", model, "Name", chass.Name,
		"PartNumber", chass.PartNumber, "SerialNumber", chass.SerialNumber,
		"BiosVersion", biosVersion, "led", chass.IndicatorLED) //nolint:staticcheck
	return &api.Board{
		VendorString:  manufacturer,
		Model:         model,
		PartNumber:    chass.PartNumber,
		SerialNumber:  chass.SerialNumber,
		BiosVersion:   biosVersion,
		IndicatorLED:  toMetalLEDState(chass.IndicatorLED), //nolint:staticcheck
		PowerMetric:   powerMetric,
		PowerSupplies: powerSupplies,
	}, nil
}

func toMetalLEDState(state schemas.IndicatorLED) string {
	switch state {
	case schemas.BlinkingIndicatorLED, schemas.LitIndicatorLED:
		return "LED-ON"
	case schemas.OffIndicatorLED, schemas.UnknownIndicatorLED:
		return "LED-OFF"
	default:
		return "LED-OFF"
	}
}

func (c *APIClient) MachineUUID() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.connectionTimeout)
	defer cancel()

	system, err := c.getSystem(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to detect machine UUID: %w", err)
	}
	if system.UUID == "" {
		return "", fmt.Errorf("machine UUID is empty")
	}
	return system.UUID, nil
}

func (c *APIClient) PowerState() (hal.PowerState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.connectionTimeout)
	defer cancel()

	system, err := c.getSystem(ctx)
	if err != nil {
		c.log.Warnw("ignore system query", "error", err)
		return hal.PowerUnknownState, nil
	}
	if system.PowerState == "" {
		return hal.PowerUnknownState, nil
	}
	return hal.GuessPowerState(string(system.PowerState)), nil
}

func (c *APIClient) PowerOn() error {
	return c.setPower(schemas.ForceOnResetType)
}

func (c *APIClient) PowerOff() error {
	return c.setPower(schemas.ForceOffResetType)
}

func (c *APIClient) PowerReset() error {
	return c.setPower(schemas.ForceRestartResetType)
}

func (c *APIClient) PowerCycle() error {
	return c.setPower(schemas.PowerCycleResetType)
}

func (c *APIClient) setPower(resetType schemas.ResetType) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.connectionTimeout)
	defer cancel()

	system, err := c.getSystem(ctx)
	if err != nil {
		return fmt.Errorf("failed to get system for power action %s: %w", resetType, err)
	}
	if _, err = system.Reset(resetType); err != nil {
		return fmt.Errorf("failed to set power to %s: %w", resetType, err)
	}
	return nil
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
	return c.setChassisIndicatorLED(schemas.LitIndicatorLED)
}

// SetChassisIdentifyLEDOff turns off the chassis identify LED
func (c *APIClient) SetChassisIdentifyLEDOff() error {
	return c.setChassisIndicatorLED(schemas.OffIndicatorLED)
}

// setChassisIndicatorLED is the shared implementation for setting the chassis indicator LED state.
func (c *APIClient) setChassisIndicatorLED(state schemas.IndicatorLED) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.connectionTimeout)
	defer cancel()

	chassis, err := c.getChassis(ctx)
	if err != nil {
		return err
	}

	payload := indicatorLEDRequest{IndicatorLED: state}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("%s/Chassis/%s", c.urlPrefix, chassis.ID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.addHeadersAndAuth(req)

	resp, err := c.doWithETag(req)
	if err != nil {
		return fmt.Errorf("unable to set chassis identify LED to %s: %w", state, err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to set chassis identify LED to %s, status code: %d", state, resp.StatusCode)
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
		Boot: schemas.Boot{
			BootSourceOverrideEnabled: schemas.ContinuousBootSourceOverrideEnabled,
			BootSourceOverrideMode:    schemas.UEFIBootSourceOverrideMode,
			BootSourceOverrideTarget:  schemas.PxeBootSource,
		},
	}
	return c.setBootTargetOverride(payload)
}

func (c *APIClient) setPersistentHDD() error {
	payload := bootOverrideRequest{
		Boot: schemas.Boot{
			BootSourceOverrideEnabled: schemas.ContinuousBootSourceOverrideEnabled,
			BootSourceOverrideMode:    schemas.UEFIBootSourceOverrideMode,
			BootSourceOverrideTarget:  schemas.HddBootSource,
		},
	}
	return c.setBootTargetOverride(payload)
}

func (c *APIClient) setBootTargetOverride(payload bootOverrideRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.connectionTimeout)
	defer cancel()

	system, err := c.getSystem(ctx)
	if err != nil {
		return err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("%s/Systems/%s", c.urlPrefix, system.ID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.addHeadersAndAuth(req)

	resp, err := c.doWithETag(req)
	if err != nil {
		return fmt.Errorf("unable to override boot order %w", err)
	}
	// Drain the body to ensure the connection can be reused
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to override boot order, http status: %s", resp.Status)
	}
	return nil
}

func (c *APIClient) addHeadersAndAuth(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.user, c.password)
}

func (c *APIClient) setNextBootBIOS() error {
	payload := bootOverrideRequest{
		Boot: schemas.Boot{
			BootSourceOverrideEnabled: schemas.OnceBootSourceOverrideEnabled,
			BootSourceOverrideMode:    schemas.UEFIBootSourceOverrideMode,
			BootSourceOverrideTarget:  schemas.BiosSetupBootSource,
		},
	}
	return c.setBootTargetOverride(payload)
}

func (c *APIClient) BMC() (*api.BMC, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.connectionTimeout)
	defer cancel()

	bmc := &api.BMC{}

	system, err := c.getSystem(ctx)
	if err != nil {
		c.log.Warnw("ignore system query", "error", err)
	} else {
		bmc.ProductManufacturer = system.Manufacturer
		bmc.ProductPartNumber = system.PartNumber
		bmc.ProductSerial = system.SerialNumber
	}

	chass, err := c.getChassis(ctx)
	if err != nil {
		c.log.Warnw("ignore chassis query", "error", err)
	} else {
		bmc.ChassisPartNumber = chass.PartNumber
		bmc.ChassisPartSerial = chass.SerialNumber
		bmc.BoardMfg = chass.Manufacturer
	}

	//TODO find bmc.BoardMfgSerial and bmc.BoardPartNumber

	return bmc, nil
}

func (c *APIClient) GetBootOptions() ([]*schemas.BootOption, error) {
	// The curl command here would be curl -k -u <user>:<pwd> https://10.1.1.18/redfish/v1/Systems/System.Embedded.1/BootOptions
	ctx, cancel := context.WithTimeout(context.Background(), c.connectionTimeout)
	defer cancel()

	system, err := c.getSystem(ctx)
	if err != nil {
		return nil, err
	}

	bootOptions, err := system.BootOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to get boot options: %w", err)
	}
	if len(bootOptions) == 0 {
		return nil, fmt.Errorf("no boot options found")
	}
	if len(system.Boot.BootOrder) == 0 {
		c.log.Warnw("no boot order found")
	}
	return bootOptions, nil
}

// SetBootOrder sets the boot order to match the sequence of the boot option entries
func (c *APIClient) SetBootOrder(entries []*schemas.BootOption) error {
	if len(entries) == 0 {
		return fmt.Errorf("cannot set boot order: no boot entries provided")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.connectionTimeout)
	defer cancel()

	system, err := c.getSystem(ctx)
	if err != nil {
		return err
	}

	if len(system.Boot.BootOrder) == 0 {
		c.log.Errorw("no boot order found")
	}

	var bootOrder []string
	for _, entry := range entries {
		bootOrder = append(bootOrder, entry.ID)
	}

	payload := bootOrderSetRequest{}
	payload.Boot.BootOrder = bootOrder

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("%s/Systems/%s", c.urlPrefix, system.ID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.addHeadersAndAuth(req)

	resp, err := c.doWithETag(req)
	if err != nil {
		return fmt.Errorf("unable to set boot order: %w", err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to set boot order, http status: %s", resp.Status)
	}
	return nil
}

// UpdateFirmware triggers a firmware update using the given URL
// BMC analyzes the file and chooses the right component to update
func (c *APIClient) UpdateFirmware(url string) error {
	// TODO NEEDS TESTING !!!
	updateURL := c.urlPrefix + "/UpdateService/Actions/UpdateService.SimpleUpdate"

	payload := struct {
		ImageURI string `json:"ImageURI,omitempty"`
	}{
		ImageURI: url,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, updateURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.addHeadersAndAuth(req)

	resp, err := c.doWithETag(req)
	if err != nil {
		return fmt.Errorf("unable to trigger update: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.log.Warnw("unable to close response body", "error", closeErr)
		}
	}()

	body, _ = io.ReadAll(resp.Body)
	// The response code is 202 for accepted, and we normally get no body
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update failed with status %s: %s", resp.Status, string(body))
	}
	c.log.Infow("update triggered successfully", "response", string(body))
	return nil
}

func (c *APIClient) getETag(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	c.addHeadersAndAuth(req)

	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	// Drain and close the body to ensure the connection can be reused
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.log.Warnw("failed to get etag, defaulting to wildcard", "status", resp.StatusCode, "url", url)
		return "*", nil
	}

	etag := resp.Header.Get("ETag")
	if etag == "" {
		return "*", nil
	}
	return etag, nil
}

func (c *APIClient) doWithETag(req *http.Request) (*http.Response, error) {
	if c.ETagRequired {
		// Create a context with timeout for the ETag fetch
		ctx := req.Context()
		if _, hasDeadline := ctx.Deadline(); !hasDeadline {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, c.connectionTimeout)
			defer cancel()
		}

		etag, err := c.getETag(ctx, req.URL.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get ETag: %w", err)
		}

		req.Header.Set("If-Match", etag)
	} else {
		req.Header.Set("If-Match", "*")
	}
	return c.Do(req)
}
