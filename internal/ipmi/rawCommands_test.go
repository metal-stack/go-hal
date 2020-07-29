package ipmi

import (
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/ipmi/ipmi"
	"github.com/metal-stack/go-hal/pkg/api"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRawCommands(t *testing.T) {
	// given
	channelNumber := uint8(1)
	uid := uint8(2)
	username := "test"
	userPassword := "secret"

	// then
	require.Equal(t, []string{"raw", "6", "67", "1", "2", "4"}, RawUserAccess(channelNumber, uid, ipmi.AdministratorPrivilege))
	require.Equal(t, []string{"raw", "6", "76", "1", "2", "2", "0", "0", "0"}, RawEnableUserSOLPayloadAccess(channelNumber, uid))
	require.Equal(t, []string{"raw", "6", "69", "2", "116", "101", "115", "116", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0"}, RawSetUserName(uid, username))
	require.Equal(t, []string{"raw", "6", "71", "2", "0"}, RawDisableUser(uid))
	require.Equal(t, []string{"raw", "6", "71", "2", "1"}, RawEnableUser(uid))
	require.Equal(t, []string{"raw", "6", "71", "130", "2", "115", "101", "99", "114", "101", "116", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0"}, RawSetUserPassword(uid, userPassword))

	require.Equal(t, []string{"raw", "0", "8", "5", "224", "4", "0", "0", "0"}, RawSetSystemBootOptions(hal.BootTargetPXE, api.VendorLenovo))
	require.Equal(t, []string{"raw", "0", "8", "5", "224", "8", "0", "0", "0"}, RawSetSystemBootOptions(hal.BootTargetDisk, api.VendorLenovo))
	require.Equal(t, []string{"raw", "0", "8", "5", "160", "24", "0", "0", "0"}, RawSetSystemBootOptions(hal.BootTargetBIOS, api.VendorLenovo))

	require.Equal(t, []string{"raw", "0", "8", "5", "224", "4", "0", "0", "0"}, RawSetSystemBootOptions(hal.BootTargetPXE, api.VendorSupermicro))
	require.Equal(t, []string{"raw", "0", "8", "5", "224", "36", "0", "0", "0"}, RawSetSystemBootOptions(hal.BootTargetDisk, api.VendorSupermicro))
	require.Equal(t, []string{"raw", "0", "8", "5", "160", "24", "0", "0", "0"}, RawSetSystemBootOptions(hal.BootTargetBIOS, api.VendorSupermicro))

	require.Equal(t, []string{"raw", "0", "8", "5", "224", "4", "0", "0", "0"}, RawSetSystemBootOptions(hal.BootTargetPXE, api.VendorVagrant))
	require.Equal(t, []string{"raw", "0", "8", "5", "224", "8", "0", "0", "0"}, RawSetSystemBootOptions(hal.BootTargetDisk, api.VendorVagrant))
	require.Equal(t, []string{"raw", "0", "8", "5", "160", "24", "0", "0", "0"}, RawSetSystemBootOptions(hal.BootTargetBIOS, api.VendorVagrant))

	require.Equal(t, []string{"raw", "0", "2", "1"}, RawChassisControl(ChassisControlPowerUp))
	require.Equal(t, []string{"raw", "0", "2", "3"}, RawChassisControl(ChassisControlHardReset))
	require.Equal(t, []string{"raw", "0", "2", "2"}, RawChassisControl(ChassisControlPowerCycle))
	require.Equal(t, []string{"raw", "0", "4", "0", "0"}, RawChassisIdentifyOff())
	require.Equal(t, []string{"raw", "0", "4", "0", "1"}, RawChassisIdentifyOn())
}
