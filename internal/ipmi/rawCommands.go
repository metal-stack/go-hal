package ipmi

import (
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/internal/ipmi/ipmi"
	"github.com/metal-stack/go-hal/pkg/api"
	"strconv"
)

func RawUserAccess(channelNumber, uid uint8, privilege ipmi.Privilege) []string {
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
	args := []uint8{AppNetworkFunction, SetUserPassword, setBit(uid, 7), 2}
	args = append(args, fixedBytes(password, 20)...)
	return rawCommand(args...)
}

func RawSetSystemBootOptions(target hal.BootTarget, vendor api.Vendor) []string {
	uefiQualifier, bootDevQualifier := GetBootOrderQualifiers(target, vendor)
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
		uu[i+1] = strconv.Itoa(int(b))
	}
	return uu
}
