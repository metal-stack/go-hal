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
}

func ViaRedfish(r *redfish.APIClient, board *api.Board, ip string, ipmiPort int, user, password string) *OutBand {
	return &OutBand{
		Redfish:  r,
		board:    board,
		ip:       ip,
		ipmiPort: ipmiPort,
		user:     user,
		password: password,
	}
}

func ViaIpmitool(ipmiTool ipmi.IpmiTool, board *api.Board, ip string, ipmiPort int, user, password string) *OutBand {
	return &OutBand{
		IpmiTool: ipmiTool,
		board:    board,
		ip:       ip,
		ipmiPort: ipmiPort,
		user:     user,
		password: password,
	}
}

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
	defer client.Close()
	return f(client)
}
