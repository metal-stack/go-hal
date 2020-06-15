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
	return rawCommand(AppNetworkFunction, SetUserAccess, channelNumber, uid, privilege)
}

func RawEnableUserSOLPayloadAccess(channelNumber, uid uint8) []string {
	return rawCommand(AppNetworkFunction, SetUserPayloadAccess, channelNumber, uid, 2, 0, 0, 0)
}

func RawSetUserName(uid uint8, username string) []string {
	args := []uint8{AppNetworkFunction, SetUserName, uid}
	args = append(args, fixedBytes(username, 16)...)
	return rawCommand(args...)
}

func RawDisableUser(uid uint8) []string {
	return rawCommand(AppNetworkFunction, SetUserPassword, uid, 0)
}

func RawEnableUser(uid uint8) []string {
	return rawCommand(AppNetworkFunction, SetUserPassword, uid, 1)
}

func RawSetUserPassword(uid uint8, password string) []string {
	args := []uint8{AppNetworkFunction, SetUserPassword, 128 + uid, 2}
	args = append(args, fixedBytes(password, 20)...)
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

func fixedBytes(s string, length int) []uint8 {
	bb := []byte(s)
	for i := len(bb); i < length; i++ {
		bb[i] = 0
	}
	if len(bb) > length {
		bb = bb[:length]
	}
	return bb
}
