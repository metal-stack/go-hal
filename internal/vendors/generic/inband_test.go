package generic

import (
	"os/exec"
	"testing"

	"github.com/gliderlabs/ssh"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/inband"
	"github.com/metal-stack/go-hal/internal/ipmi"
	"github.com/metal-stack/go-hal/pkg/api"
)

type mockIpmiTool struct {
	devicePresent bool
	userListOut   string
	runArgs       [][]string
	bmc           *api.BMC
}

func (m *mockIpmiTool) DevicePresent() bool { return m.devicePresent }
func (m *mockIpmiTool) NewCommand(arg ...string) (*exec.Cmd, error) {
	return nil, nil
}
func (m *mockIpmiTool) Run(arg ...string) (string, error) {
	m.runArgs = append(m.runArgs, arg)
	if len(arg) >= 3 && arg[0] == "user" && arg[1] == "list" {
		return m.userListOut, nil
	}
	return "", nil
}
func (m *mockIpmiTool) CreateUser(user api.BMCUser, privilege api.IpmiPrivilege, password string, constraints *api.PasswordConstraints, apiType ipmi.ApiType) (string, error) {
	return "generated-password", nil
}
func (m *mockIpmiTool) ChangePassword(user api.BMCUser, newPassword string, apiType ipmi.ApiType) error {
	return nil
}
func (m *mockIpmiTool) NeedsPasswordChange(user api.BMCUser, password string) (bool, error) {
	return true, nil
}
func (m *mockIpmiTool) SetUserEnabled(user api.BMCUser, enabled bool, apiType ipmi.ApiType) error {
	return nil
}
func (m *mockIpmiTool) GetLanConfig() (ipmi.LanConfig, error) {
	return ipmi.LanConfig{}, nil
}
func (m *mockIpmiTool) SetBootOrder(target hal.BootTarget, vendor api.Vendor) error {
	return nil
}
func (m *mockIpmiTool) SetChassisControl(ipmi.ChassisControlFunction) error {
	return nil
}
func (m *mockIpmiTool) SetChassisIdentifyLEDState(hal.IdentifyLEDState) error {
	return nil
}
func (m *mockIpmiTool) SetChassisIdentifyLEDOn() error {
	return nil
}
func (m *mockIpmiTool) SetChassisIdentifyLEDOff() error {
	return nil
}
func (m *mockIpmiTool) GetFru() (ipmi.Fru, error) {
	return ipmi.Fru{}, nil
}
func (m *mockIpmiTool) GetSession() (ipmi.Session, error) {
	return ipmi.Session{}, nil
}
func (m *mockIpmiTool) BMC() (*api.BMC, error) {
	return m.bmc, nil
}
func (m *mockIpmiTool) MachineUUID() (string, error) {
	return "", nil
}
func (m *mockIpmiTool) OpenConsole(s ssh.Session) error {
	return nil
}

func Test_parseUsers(t *testing.T) {
	input := `ID  Name       Callin  Link Auth  IPMI Msg  Channel Priv Limit
1   ADMIN      false   false      true       ADMINISTRATOR
2   USERID     false   false      true       ADMINISTRATOR
3   metal      false   false      true       ADMINISTRATOR
`
	users := parseUsers(input)
	if len(users) != 3 {
		t.Fatalf("expected 3 users, got %d: %#v", len(users), users)
	}
	if users[0].Name != "ADMIN" || users[0].Id != "1" {
		t.Errorf("unexpected first user: %#v", users[0])
	}
	if users[1].Name != "USERID" || users[1].Id != "2" {
		t.Errorf("unexpected second user: %#v", users[1])
	}
}

func Test_bmcConnection_Present(t *testing.T) {
	conn := newBMCConnection(&mockIpmiTool{devicePresent: true})
	if !conn.Present() {
		t.Errorf("expected Present() to be true")
	}
	conn = newBMCConnection(&mockIpmiTool{devicePresent: false})
	if conn.Present() {
		t.Errorf("expected Present() to be false")
	}
}

func Test_bmcConnection_PresentSuperUser(t *testing.T) {
	conn := newBMCConnection(&mockIpmiTool{userListOut: `ID  Name    Callin  Link Auth  IPMI Msg  Channel Priv Limit
1   ADMIN   false   false      true       ADMINISTRATOR
`})
	u := conn.PresentSuperUser()
	if u.Name != "ADMIN" || u.Id != "1" {
		t.Errorf("unexpected superuser: %#v", u)
	}
}

func Test_bmcConnection_PresentSuperUser_fallback(t *testing.T) {
	conn := newBMCConnection(&mockIpmiTool{userListOut: ""})
	u := conn.PresentSuperUser()
	if u.Name != "ADMIN" || u.Id != "1" {
		t.Errorf("unexpected fallback superuser: %#v", u)
	}
}

func Test_bmcConnection_User_picksFreeSlot(t *testing.T) {
	// ids 2 and 3 are taken, so a free slot (4) must be chosen.
	conn := newBMCConnection(&mockIpmiTool{userListOut: `ID  Name    Callin  Link Auth  IPMI Msg  Channel Priv Limit
1   ADMIN   false   false      true       ADMINISTRATOR
2   USERID  false   false      true       ADMINISTRATOR
3   metal   false   false      true       ADMINISTRATOR
`})
	u := conn.User()
	if u.Id != "4" {
		t.Errorf("expected free slot 4, got %#v", u)
	}
	if u.Name != "metal" {
		t.Errorf("expected user name metal, got %s", u.Name)
	}
}

func Test_bmcConnection_User_defaultID(t *testing.T) {
	// only id 1 is taken, id 2 is free.
	conn := newBMCConnection(&mockIpmiTool{userListOut: `ID  Name    Callin  Link Auth  IPMI Msg  Channel Priv Limit
1   ADMIN   false   false      true       ADMINISTRATOR
`})
	u := conn.User()
	if u.Id != "2" {
		t.Errorf("expected first free slot 2, got %#v", u)
	}
}

func newBMCConnection(tool ipmi.IpmiTool) *bmcConnection {
	ib := &inBand{
		InBand: &inband.InBand{
			IpmiTool: tool,
		},
	}
	return &bmcConnection{
		inBand: ib,
	}
}
