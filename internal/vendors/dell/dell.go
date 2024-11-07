package dell

import (
	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/outband"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/metal-stack/go-hal/pkg/logger"
)

type (
	outBand struct {
		*outband.OutBand
		log logger.Logger
	}
	bmcConnectionOutBand struct {
		*outBand
	}
)

// OutBand creates an outband connection to a supermicro server.
func OutBand(r *redfish.APIClient, board *api.Board, ip string, ipmiPort int, user, password string, log logger.Logger) (hal.OutBand, error) {
	i, err := ipmi.NewOutBand(ip, ipmiPort, user, password, log)
	if err != nil {
		return nil, err
	}
	return &outBand{
		OutBand: outband.New(r, i, board, ip, ipmiPort, user, password),
		log:     log,
	}, nil
}

// BMCConnection implements hal.OutBand.
func (ob *outBand) BMCConnection() api.OutBandBMCConnection {
	return &bmcConnectionOutBand{
		outBand: ob,
	}
}

func (c *bmcConnectionOutBand) BMC() (*api.BMC, error) {
	return c.IpmiTool.BMC()
}

// BootFrom implements hal.OutBand.
func (o *outBand) BootFrom(hal.BootTarget) error {
	panic("unimplemented")
}

// Console implements hal.OutBand.
func (o *outBand) Console(ssh.Session) error {
	panic("unimplemented")
}

// Describe implements hal.OutBand.
func (o *outBand) Describe() string {
	panic("unimplemented")
}

// IPMIConnection implements hal.OutBand.
// Subtle: this method shadows the method (*OutBand).IPMIConnection of outBand.OutBand.
func (o *outBand) IPMIConnection() (ip string, port int, user string, password string) {
	panic("unimplemented")
}

// IdentifyLEDOff implements hal.OutBand.
func (o *outBand) IdentifyLEDOff() error {
	panic("unimplemented")
}

// IdentifyLEDOn implements hal.OutBand.
func (o *outBand) IdentifyLEDOn() error {
	panic("unimplemented")
}

// IdentifyLEDState implements hal.OutBand.
func (o *outBand) IdentifyLEDState(hal.IdentifyLEDState) error {
	panic("unimplemented")
}

// PowerCycle implements hal.OutBand.
func (o *outBand) PowerCycle() error {
	panic("unimplemented")
}

// PowerOff implements hal.OutBand.
func (o *outBand) PowerOff() error {
	panic("unimplemented")
}

// PowerOn implements hal.OutBand.
func (o *outBand) PowerOn() error {
	panic("unimplemented")
}

// PowerReset implements hal.OutBand.
func (o *outBand) PowerReset() error {
	panic("unimplemented")
}

// PowerState implements hal.OutBand.
func (o *outBand) PowerState() (hal.PowerState, error) {
	panic("unimplemented")
}

// UUID implements hal.OutBand.
func (o *outBand) UUID() (*uuid.UUID, error) {
	panic("unimplemented")
}

// UpdateBIOS implements hal.OutBand.
func (o *outBand) UpdateBIOS(url string) error {
	panic("unimplemented")
}

// UpdateBMC implements hal.OutBand.
func (o *outBand) UpdateBMC(url string) error {
	panic("unimplemented")
}
