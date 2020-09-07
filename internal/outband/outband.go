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

func New(r *redfish.APIClient, board *api.Board, ip string, ipmiPort int, user, password string) (*OutBand, error) {
	i, err := ipmi.New()
	if err != nil {
		return nil, err
	}

	return &OutBand{
		Redfish:  r,
		IpmiTool: i,
		board:    board,
		ip:       ip,
		ipmiPort: ipmiPort,
		user:     user,
		password: password,
	}, nil
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
