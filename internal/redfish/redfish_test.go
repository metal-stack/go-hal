package redfish_test

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/metal-stack/go-hal/pkg/logger"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var connectionTimeout = pointer.Pointer(10 * time.Second)

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

	tests := []struct {
		name     string
		url      string
		user     string
		password string
		insecure bool
		log      logger.Logger
		state    hal.IdentifyLEDState
		wantErr  error
	}{
		{
			name:     "ledstate",
			url:      endpoint,
			log:      logger.NewSlog(log),
			state:    hal.IdentifyLEDStateOff,
			insecure: true,
			wantErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := redfish.New(tt.url, tt.user, tt.password, tt.insecure, tt.log, connectionTimeout)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			err = c.SetChassisIdentifyLEDState(tt.state)
			if diff := cmp.Diff(err, tt.wantErr); diff != "" {
				t.Errorf("diff = %s", diff)
			}
		})
	}
}

func TestAPIClient_setPower(t *testing.T) {
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
		state    hal.IdentifyLEDState
		wantErr  error
	}{
		{
			name:     "power",
			url:      endpoint,
			log:      logger.NewSlog(log),
			state:    hal.IdentifyLEDStateOff,
			insecure: true,
			wantErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := redfish.New(tt.url, tt.user, tt.password, tt.insecure, tt.log, connectionTimeout)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			err = c.SetChassisIdentifyLEDState(tt.state)
			if diff := cmp.Diff(err, tt.wantErr); diff != "" {
				t.Errorf("diff = %s", diff)
			}
		})
	}
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
