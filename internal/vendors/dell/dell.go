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

func (ob *outBand) Close() {
	ob.Redfish.Gofish.Logout()
}

func (c *bmcConnectionOutBand) BMC() (*api.BMC, error) {
	board, err := c.outBand.Redfish.BoardInfo()
	if err != nil {
		return nil, err
	}
	return &api.BMC{
		IP:                  c.outBand.Ip,
		MAC:                 "",
		ChassisPartNumber:   "",
		ChassisPartSerial:   "",
		BoardMfg:            "",
		BoardMfgSerial:      "",
		BoardPartNumber:     board.PartNumber,
		ProductManufacturer: board.VendorString,
		ProductPartNumber:   "",
		ProductSerial:       board.SerialNumber,
		FirmwareRevision:    "",
	}, nil
}

// BootFrom implements hal.OutBand.
func (o *outBand) BootFrom(hal.BootTarget) error {
	panic("unimplemented")
}

// Console implements hal.OutBand.
func (o *outBand) Console(ssh.Session) error {
	// Console access must be switched to ssh root@<IP> console com2
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

func (o *outBand) GetIdentifyLED() (hal.IdentifyLEDState, error) {
	return o.Redfish.GetIdentifyLED()
}

// IdentifyLEDOff implements hal.OutBand.
func (o *outBand) IdentifyLEDOff() error {
	return o.Redfish.IdentifyLEDState(hal.IdentifyLEDStateOff)
}

// IdentifyLEDOn implements hal.OutBand.
func (o *outBand) IdentifyLEDOn() error {
	return o.Redfish.IdentifyLEDState(hal.IdentifyLEDStateOn)
}

// IdentifyLEDState implements hal.OutBand.
func (o *outBand) IdentifyLEDState(state hal.IdentifyLEDState) error {
	return o.Redfish.IdentifyLEDState(state)
}

// PowerCycle implements hal.OutBand.
func (o *outBand) PowerCycle() error {
	return o.Redfish.PowerCycle()
}

// PowerOff implements hal.OutBand.
func (o *outBand) PowerOff() error {
	return o.Redfish.PowerOff()
}

// PowerOn implements hal.OutBand.
func (o *outBand) PowerOn() error {
	return o.Redfish.PowerOn()
}

// PowerReset implements hal.OutBand.
func (o *outBand) PowerReset() error {
	return o.Redfish.PowerReset()
}

// PowerState implements hal.OutBand.
func (o *outBand) PowerState() (hal.PowerState, error) {
	return o.Redfish.PowerState()
}

// UUID implements hal.OutBand.
func (o *outBand) UUID() (*uuid.UUID, error) {
	uuidString, err := o.Redfish.MachineUUID()
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(uuidString)
	return &id, err
}

// UpdateBIOS implements hal.OutBand.
func (o *outBand) UpdateBIOS(url string) error {
	panic("unimplemented")
}

// UpdateBMC implements hal.OutBand.
func (o *outBand) UpdateBMC(url string) error {
	panic("unimplemented")
}
