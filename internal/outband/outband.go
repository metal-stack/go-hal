package outband

import (
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/pkg/api"
)

type OutBand struct {
	Redfish  *redfish.APIClient
	IpmiTool ipmi.IpmiTool
	board    *api.Board
	ip       string
	ipmiPort int
	user     string
	password string
	sshPort  int
}

// ViaRedfish returns an out-band connection that uses the given redfish client
func ViaRedfish(r *redfish.APIClient, board *api.Board) *OutBand {
	return &OutBand{
		Redfish: r,
		board:   board,
	}
}

// ViaRedfishPlusSSH returns an out-band connection that uses the given redfish client plus saves ssh connection data
func ViaRedfishPlusSSH(r *redfish.APIClient, board *api.Board, user, password, ip string, sshPort int) *OutBand {
	return &OutBand{
		Redfish:  r,
		board:    board,
		user:     user,
		password: password,
		ip:       ip,
		sshPort:  sshPort,
	}
}

// New returns an out-band connection that uses the given redfish client and ipmitool as well as a goipmi client
func New(r *redfish.APIClient, ipmiTool ipmi.IpmiTool, board *api.Board, ip string, ipmiPort int, user, password string) *OutBand {
	return &OutBand{
		Redfish:  r,
		IpmiTool: ipmiTool,
		board:    board,
		ip:       ip,
		ipmiPort: ipmiPort,
		user:     user,
		password: password,
	}
}

// ViaGoipmi returns an out-band connection that uses a goipmi client
func ViaGoipmi(board *api.Board, ip string, ipmiPort int, user, password string) *OutBand {
	return &OutBand{
		board:    board,
		ip:       ip,
		ipmiPort: ipmiPort,
		user:     user,
		password: password,
	}
}

func (ob *OutBand) Board() *api.Board {
	return ob.board
}

func (ob *OutBand) IPMIConnection() (string, int, string, string) {
	return ob.ip, ob.ipmiPort, ob.user, ob.password
}

func (ob *OutBand) Goipmi(f func(*ipmi.Client) error) error {
	client, err := ipmi.OpenClientConnection(ob.IPMIConnection())
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close()
	}()

	return f(client)
}

func (ob *OutBand) GetUsername() string {
	return ob.user
}

func (ob *OutBand) GetPassword() string {
	return ob.password
}

func (ob *OutBand) GetIP() string {
	return ob.ip
}

func (ob *OutBand) GetSSHPort() int {
	return ob.sshPort
}
