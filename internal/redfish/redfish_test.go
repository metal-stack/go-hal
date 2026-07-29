package redfish_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/metal-stack/go-hal/pkg/logger"
	"github.com/stmcginnis/gofish/schemas"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var connectionTimeout = new(10 * time.Second)

func TestAPIClient_BoardInfo(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	endpoint, err := startRedfishMock(t)
	require.NoError(t, err)

	tests := []struct {
		name     string
		url      string
		user     string
		password string
		insecure bool
		log      logger.Logger
		want     *api.Board
		wantErr  error
	}{
		{
			name:     "boardinfo",
			url:      endpoint,
			log:      logger.NewSlog(log),
			insecure: true,
			want: &api.Board{
				VM:           false,
				VendorString: "Contoso",
				Model:        "3500",
				PartNumber:   "224071-J23",
				SerialNumber: "437XR1138R2",
				BiosVersion:  "P79 v1.45 (12/06/2017)",
				IndicatorLED: "LED-ON",
				PowerMetric: &api.PowerMetric{
					AverageConsumedWatts: 319,
					IntervalInMin:        30,
					MaxConsumedWatts:     489,
					MinConsumedWatts:     271,
				},
				PowerSupplies: []api.PowerSupply{
					{Status: api.Status{Health: "Warning", State: "Enabled"}},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := redfish.New(tt.url, tt.user, tt.password, tt.insecure, tt.log, connectionTimeout)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got, err := c.BoardInfo()
			if diff := cmp.Diff(err, tt.wantErr); diff != "" {
				t.Errorf("diff = %s", diff)
			}
			if diff := cmp.Diff(
				tt.want, got); diff != "" {
				t.Errorf("machineServiceServer.BMCCommand() = %v, want %v diff: %s", got, tt.want, diff)
			}
		})
	}
}

func TestAPIClient_MachineUUID(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	endpoint, err := startRedfishMock(t)
	require.NoError(t, err)

	tests := []struct {
		name     string
		url      string
		user     string
		password string
		insecure bool
		log      logger.Logger
		want     string
		wantErr  error
	}{
		{
			name:     "machine uuid",
			url:      endpoint,
			log:      logger.NewSlog(log),
			insecure: true,
			want:     "38947555-7742-3448-3784-823347823834",
			wantErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := redfish.New(tt.url, tt.user, tt.password, tt.insecure, tt.log, connectionTimeout)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got, err := c.MachineUUID()
			if diff := cmp.Diff(err, tt.wantErr); diff != "" {
				t.Errorf("diff = %s", diff)
			}
			if diff := cmp.Diff(
				tt.want, got); diff != "" {
				t.Errorf("machineServiceServer.BMCCommand() = %v, want %v diff: %s", got, tt.want, diff)
			}
		})
	}
}

func TestAPIClient_SetChassisIdentifyLEDState(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	endpoint, err := startRedfishMock(t)
	require.NoError(t, err)

	chassis := getChassisData(t, endpoint)
	require.NotContains(t, chassis, "LocationIndicatorActive", "legacy path must not set to test the legacy IdentifyLEDState property")

	tests := []struct {
		name     string
		user     string
		password string
		log      logger.Logger
		state    hal.IdentifyLEDState
		wantErr  error
	}{
		{
			name:    "Off",
			log:     logger.NewSlog(log),
			state:   hal.IdentifyLEDStateOff,
			wantErr: nil,
		},
		{
			name:    "On",
			log:     logger.NewSlog(log),
			state:   hal.IdentifyLEDStateOn,
			wantErr: nil,
		},
		{
			name:    "Unknown",
			log:     logger.NewSlog(log),
			state:   hal.IdentifyLEDStateUnknown,
			wantErr: fmt.Errorf("unknown identify LED state: UNKNOWN"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := redfish.New(endpoint, tt.user, tt.password, true, tt.log, connectionTimeout)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			err = c.SetChassisIdentifyLEDState(tt.state)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				return
			}
			require.NoError(t, err)

			chassis := getChassisData(t, endpoint)
			require.NotContains(t, chassis, "LocationIndicatorActive", "legacy path must not set LocationIndicatorActive")

			value, ok := chassis["IndicatorLED"]
			require.True(t, ok)

			var want schemas.IndicatorLED
			switch tt.state {
			case hal.IdentifyLEDStateOff:
				want = schemas.OffIndicatorLED
			case hal.IdentifyLEDStateOn:
				want = schemas.LitIndicatorLED
			}

			require.EqualValues(t, want, value)
		})
	}
}

func TestAPIClient_SetLocationIndicatorActive(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	endpoint, err := startRedfishMock(t)
	require.NoError(t, err)
	resourcePath := "/redfish/v1/Chassis/1U"

	tests := []struct {
		name                         string
		user                         string
		password                     string
		log                          logger.Logger
		targetState                  hal.IdentifyLEDState
		givenLocationIndicatorActive bool
		wantErr                      error
	}{
		{
			name:                         "off-to-on",
			log:                          logger.NewSlog(log),
			targetState:                  hal.IdentifyLEDStateOn,
			givenLocationIndicatorActive: false,
			wantErr:                      nil,
		},
		{
			name:                         "off-to-off",
			log:                          logger.NewSlog(log),
			targetState:                  hal.IdentifyLEDStateOff,
			givenLocationIndicatorActive: false,
			wantErr:                      nil,
		},
		{
			name:                         "on-to-on",
			log:                          logger.NewSlog(log),
			targetState:                  hal.IdentifyLEDStateOn,
			givenLocationIndicatorActive: true,
			wantErr:                      nil,
		},
		{
			name:                         "on-to-off",
			log:                          logger.NewSlog(log),
			targetState:                  hal.IdentifyLEDStateOff,
			givenLocationIndicatorActive: true,
			wantErr:                      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = patchRedfishMockSchema(endpoint, resourcePath, locationIndicatorActiveStruct{LocationIndicatorActive: tt.givenLocationIndicatorActive})
			require.NoError(t, err)

			c, err := redfish.New(endpoint, tt.user, tt.password, true, tt.log, connectionTimeout)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			err = c.SetChassisIdentifyLEDState(tt.targetState)
			if diff := cmp.Diff(err, tt.wantErr); diff != "" {
				t.Errorf("diff = %s", diff)
			}
			chassis := getChassisData(t, endpoint)
			switch tt.targetState {
			case hal.IdentifyLEDStateOff:
				require.False(t, getBool(t, chassis, "LocationIndicatorActive"), "OFF must set LocationIndicatorActive=false")
			case hal.IdentifyLEDStateOn:
				require.True(t, getBool(t, chassis, "LocationIndicatorActive"), "ON must set LocationIndicatorActive=true")
			}
		})
	}
}

func getChassisData(t *testing.T, endpoint string) map[string]any {
	t.Helper()
	return getRedfishData(t, endpoint+"/redfish/v1/Chassis/1U")
}

func getBool(t *testing.T, data map[string]any, key string) bool {
	t.Helper()

	_, ok := data[key]
	require.True(t, ok)

	return data[key].(bool)
}

func getRedfishData(t *testing.T, url string) map[string]any {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: *connectionTimeout}
	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()
	require.Falsef(t, resp.StatusCode < 200 || resp.StatusCode >= 300, "get returned status %d", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var data map[string]any
	err = json.Unmarshal(body, &data)
	require.NoError(t, err)

	return data
}

type locationIndicatorActiveStruct struct {
	LocationIndicatorActive bool `json:"LocationIndicatorActive"`
}

func patchRedfishMockSchema(endpoint, resourcePath string, patch any) error {
	rawPatch, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	url := endpoint + resourcePath
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(rawPatch))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: *connectionTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("patch returned status %d", resp.StatusCode)
	}
	return nil
}

func startRedfishMock(t *testing.T) (string, error) {
	ctx := t.Context()

	c, err := testcontainers.Run(
		ctx,
		"dmtf/redfish-mockup-server:latest",
		testcontainers.WithExposedPorts("8000/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("8000/tcp").WithStartupTimeout(time.Second*5),
			wait.ForExposedPort(),
		),
	)
	require.NoError(t, err)

	endpoint, err := c.Endpoint(ctx, "http")
	return endpoint, err
}
