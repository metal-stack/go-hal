package ipmi

import (
	"fmt"
	"github.com/metal-stack/go-hal"
	"github.com/metal-stack/go-hal/pkg/api"
	goipmi "github.com/vmware/goipmi"
)

func RawSetSystemBootOptions(target hal.BootTarget, compliance api.Compliance) []string {
	uefiQualifier, bootDevQualifier := GetBootOrderQualifiers(target, compliance)
	return rawCommand(ChassisNetworkFunction, SetSystemBootOptions, goipmi.BootParamBootFlags, uefiQualifier, bootDevQualifier, 0, 0, 0)
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
