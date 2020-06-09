package ipmi

import (
	"fmt"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/pkg/api"
)

func RawSetChannelAccess(channelNumber uint8, privilege Privilege) []string {
	return rawCommand(AppNetworkFunction, SetChannelAccess, channelNumber, 0, privilege)
}

func RawUserAccess(channelNumber, uid uint8, privilege Privilege) []string {
	return rawCommand(AppNetworkFunction, SetChannelAccess, channelNumber, uid, privilege)
}

func RawActivateSOLPayload(channelNumber uint8) []string {
	return rawCommand(AppNetworkFunction, SetSOLConfigurationParameters, 1, channelNumber, 1)
}

func RawSetUserName(uid uint8, username string) []string {
	args := []uint8{AppNetworkFunction, SetUserName, uid}
	bb := []byte(username)
	if len(bb) > 16 {
		bb = bb[:16]
	}
	args = append(args, bb...)
	return rawCommand(args...)
}

func RawEnableUser(uid uint8) []string {
	return rawCommand(AppNetworkFunction, SetUserPassword, uid, 1)
}

func RawSetUserPassword(uid uint8, password string) []string {
	args := []uint8{AppNetworkFunction, SetUserPassword, uid, 2}
	bb := []byte(password)
	if len(bb) > 16 {
		bb = bb[:16]
	}
	args = append(args, bb...)
	return rawCommand(args...)
}

func RawSetSystemBootOptions(target hal.BootTarget, compliance api.Compliance) []string {
	uefiQualifier, bootDevQualifier := GetBootOrderQualifiers(target, compliance)
	return rawCommand(ChassisNetworkFunction, SetSystemBootOptions, BootFlags, uefiQualifier, bootDevQualifier, 0, 0, 0)
}

func RawChassisControl(fn ChassisControlFunction) []string {
	return rawCommand(ChassisNetworkFunction, ChassisControl, fn)
}

func RawChassisIdentifyOff() []string {
	return rawCommand(ChassisNetworkFunction, ChassisIdentify, ChassisIdentifyForceOnIndefinitely, False)
}

func RawChassisIdentifyOn() []string {
	return rawCommand(ChassisNetworkFunction, ChassisIdentify, ChassisIdentifyForceOnIndefinitely, True)
}

func rawCommand(bytes ...uint8) []string {
	uu := make([]string, len(bytes)+1)
	uu[0] = "raw"
	for i, b := range bytes {
		uu[i+1] = fmt.Sprintf("%X", b)
	}
	return uu
}
