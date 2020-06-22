package outband

import (
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/pkg/api"
)

type OutBand struct {
	Redfish  *redfish.APIClient
	board    *api.Board
	ip       string
	user     string
	password string
}

func New(r *redfish.APIClient, board *api.Board, ip, user, password string) *OutBand {
	return &OutBand{
		Redfish:  r,
		board:    board,
		ip:       ip,
		user:     user,
		password: password,
	}
}

func (ob *OutBand) Board() *api.Board {
	return ob.board
}

func (ob *OutBand) Connection() (string, string, string) {
	return ob.ip, ob.user, ob.password
}

func (ob *OutBand) Goipmi(f func(*ipmi.Client) error) error {
	client, err := ipmi.OpenClientConnection(ob.Connection())
	if err != nil {
		return err
	}
	defer client.Close()
	return f(client)
}
