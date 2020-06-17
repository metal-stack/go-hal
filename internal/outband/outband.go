package outband

import (
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/internal/redfish"
	"github.com/metal-stack/go-hal/pkg/api"
)

type OutBand struct {
	Ipmi       ipmi.Ipmi
	Redfish    *redfish.APIClient
	board      *api.Board
	compliance api.Compliance
	ip         string
	user       string
	password   string
}

func New(r *redfish.APIClient, board *api.Board, compliance api.Compliance, ip, user, password string) (*OutBand, error) {
	i, err := ipmi.New("ipmitool", compliance)
	if err != nil {
		return nil, err
	}
	return &OutBand{
		Ipmi:       i,
		Redfish:    r,
		board:      board,
		compliance: compliance,
		ip:         ip,
		user:       user,
		password:   password,
	}, nil
}

func (ob *OutBand) Board() *api.Board {
	return ob.board
}

func (ob *OutBand) Connection() (string, string, string) {
	return ob.ip, ob.user, ob.password
}
