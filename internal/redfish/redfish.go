package redfish

import (
	"bytes"
	"context"
	"encoding/base64"
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
	basicAuth         string // TODO Why do we need this? Seems to be never used
	log               logger.Logger
	connectionTimeout time.Duration
	chassisID         string
	systemID          string
}

type bootOverrideRequest struct {
	Boot schemas.Boot `json:"Boot"`
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

	apiClient := &APIClient{
		client:            c,
		Client:            c.HTTPClient,
		user:              user,
		password:          password,
		basicAuth:         base64.StdEncoding.EncodeToString([]byte(user + ":" + password)),
		urlPrefix:         fmt.Sprintf("%s/redfish/v1", url),
		log:               log,
		connectionTimeout: timeout,
		chassisID:         "",
		systemID:          "",
	}

	// Discover systemID and chassisID
	if err := apiClient.discoverIDs(ctx); err != nil {
		log.Warnw("failed to auto-discover system/chassis ID, using defaults", "error", err)
	}

	if apiClient.systemID == "" {
		apiClient.systemID = "Self" // Default SystemID to ensure backwards compatibility
	}
	if apiClient.chassisID == "" {
		apiClient.chassisID = "1" // Default ChassisID to ensure backwards compatibility
	}

	return apiClient, nil
}

// discoverIDs attempts to automatically discover the systemID and chassisID
func (c *APIClient) discoverIDs(ctx context.Context) error {
	g := c.client.WithContext(ctx)

	if g.Service == nil {
		return fmt.Errorf("gofish service root is not available")
	}

	// Discover System ID
	systems, err := g.Service.Systems()
	if err != nil {
		return fmt.Errorf("failed to query systems: %w", err)
	}

	if len(systems) > 0 {
		// Use the ID from the first system
		c.systemID = systems[0].ID
		c.log.Debugw("discovered system ID", "systemID", c.systemID)
	} else {
		c.log.Warnw("no systems found during discovery")
	}

	// Discover Chassis ID
	chassis, err := g.Service.Chassis()
	if err != nil {
		return fmt.Errorf("failed to query chassis: %w", err)
	}

	// Prefer RackMount chassis, but fall back to any chassis if none found
	var selectedChassis *schemas.Chassis
	for _, chass := range chassis {
		if chass.ChassisType == schemas.RackMountChassisType {
			selectedChassis = chass
			break
		}
	}

	// If no RackMount chassis found, use the first available
	if selectedChassis == nil && len(chassis) > 0 {
		selectedChassis = chassis[0]
		c.log.Debugw("no RackMount chassis found, using first available",
			"chassisType", selectedChassis.ChassisType)
	}

	if selectedChassis != nil {
		c.chassisID = selectedChassis.ID
		c.log.Debugw("discovered chassis ID", "chassisID", c.chassisID, "chassisType", selectedChassis.ChassisType)
	} else {
		c.log.Warnw("no chassis found during discovery")
	}

	// Error if we couldn't find either ID
	if c.systemID == "" || c.chassisID == "" {
		return fmt.Errorf("failed to discover any system or chassis IDs")
	}

	return nil
}

func (c *APIClient) SetChassisID(id string) {
	c.chassisID = id
}

func (c *APIClient) SetSystemID(id string) {
	c.systemID = id
}

func (c *APIClient) BoardInfo() (*api.Board, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.connectionTimeout)
	defer cancel()
	g := c.client.WithContext(ctx)
	// Query the chassis data using the session token
	if g.Service == nil {
		return nil, fmt.Errorf("gofish service root is not available most likely due to missing username")
	}

	biosVersion := ""
	manufacturer := ""
	model := ""

	systems, err := g.Service.Systems()
	if err != nil {
		c.log.Warnw("ignore system query", "error", err.Error())
	}
	for _, system := range systems {
		if system.BiosVersion != "" {
			biosVersion = system.BiosVersion
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

	chassis, err := g.Service.Chassis()
	if err != nil {
		c.log.Warnw("ignore system query", "error", err.Error())
	}
	for _, chass := range chassis {
		if chass.ChassisType == schemas.RackMountChassisType {
			power, err := chass.Power()
			var powerMetric *api.PowerMetric
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
	}
	return nil, fmt.Errorf("no board detected: #chassis:%d", len(chassis))
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

// MachineUUID retrieves a unique uuid for this (hardware) machine
func (c *APIClient) MachineUUID() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.connectionTimeout)
	defer cancel()
	g := c.client.WithContext(ctx)
	systems, err := g.Service.Systems()
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
	ctx, cancel := context.WithTimeout(context.Background(), c.connectionTimeout)
	defer cancel()
	g := c.client.WithContext(ctx)
	systems, err := g.Service.Systems()
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
	g := c.client.WithContext(ctx)
	systems, err := g.Service.Systems()
	if err != nil {
		c.log.Warnw("ignore system query", "error", err.Error())
	}
	for _, system := range systems {
		if _, err = system.Reset(resetType); err == nil {
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
		IndicatorLED: schemas.LitIndicatorLED,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPatch, fmt.Sprintf("%s/Chassis/%s", c.urlPrefix, c.chassisID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.addHeadersAndAuth(req)

	resp, err := c.doWithETag(req)
	if err == nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("unable to turn on the chassis identify LED %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unable to turn on the chassis identify LED, status code: %d", resp.StatusCode)
	}
	return nil
}

// SetChassisIdentifyLEDOff turns off the chassis identify LED
func (c *APIClient) SetChassisIdentifyLEDOff() error {
	payload := indicatorLEDRequest{
		IndicatorLED: schemas.OffIndicatorLED,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPatch, fmt.Sprintf("%s/Chassis/%s", c.urlPrefix, c.chassisID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.addHeadersAndAuth(req)

	resp, err := c.doWithETag(req)
	if err == nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("unable to turn off the chassis identify LED %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unable to turn off the chassis identify LED, status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *APIClient) SetBootOrder(target hal.BootTarget) error {
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
	return c.setBootOrderOverride(payload)
}

func (c *APIClient) setPersistentHDD() error {
	payload := bootOverrideRequest{
		Boot: schemas.Boot{
			BootSourceOverrideEnabled: schemas.ContinuousBootSourceOverrideEnabled,
			BootSourceOverrideMode:    schemas.UEFIBootSourceOverrideMode,
			BootSourceOverrideTarget:  schemas.HddBootSource,
		},
	}
	return c.setBootOrderOverride(payload)
}

func (c *APIClient) setBootOrderOverride(payload bootOverrideRequest) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPatch, fmt.Sprintf("%s/Systems/%s", c.urlPrefix, c.systemID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.addHeadersAndAuth(req)

	resp, err := c.doWithETag(req)
	if err == nil {
		// TODO drain body?
		_ = resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("unable to override boot order %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unable to override boot order, status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *APIClient) addHeadersAndAuth(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
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
	return c.setBootOrderOverride(payload)
}

func (c *APIClient) BMC() (*api.BMC, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.connectionTimeout)
	defer cancel()
	g := c.client.WithContext(ctx)
	systems, err := g.Service.Systems()
	if err != nil {
		c.log.Warnw("ignore system query", "error", err.Error())
	}

	chassis, err := g.Service.Chassis()
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
		if chass.ChassisType != schemas.RackMountChassisType {
			continue
		}

		bmc.ChassisPartNumber = chass.PartNumber
		bmc.ChassisPartSerial = chass.SerialNumber

		bmc.BoardMfg = chass.Manufacturer
	}

	//TODO find bmc.BoardMfgSerial and bmc.BoardPartNumber

	return bmc, nil
}

// TODO should we cache this?
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
	defer func() {
		_ = resp.Body.Close()
	}()

	// Drain the body to ensure the connection can be reused
	_, _ = io.Copy(io.Discard, resp.Body)

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
	etag, err := c.getETag(req.Context(), req.URL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get ETag: %w", err)
	}

	req.Header.Set("If-Match", etag)
	return c.Do(req)
}
