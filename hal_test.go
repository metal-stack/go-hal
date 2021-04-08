package hal

import (
	"testing"
)

func TestPowerState_String(t *testing.T) {
	tests := []struct {
		name string
		p    PowerState
		want string
	}{
		{name: "ON", p: PowerOnState, want: "ON"},
		{name: "OF", p: PowerOffState, want: "OFF"},
		{name: "UNKNOWN", p: PowerUnknownState, want: "UNKNOWN"},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.String(); got != tt.want {
				t.Errorf("PowerState.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBootTarget_String(t *testing.T) {
	tests := []struct {
		name string
		b    BootTarget
		want string
	}{
		{name: "BIOS", b: BootTargetBIOS, want: "BIOS"},
		{name: "DISK", b: BootTargetDisk, want: "DISK"},
		{name: "PXE", b: BootTargetPXE, want: "PXE"},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.String(); got != tt.want {
				t.Errorf("BootTarget.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIdentifyLEDState_String(t *testing.T) {
	tests := []struct {
		name string
		i    IdentifyLEDState
		want string
	}{
		{name: "ON", i: IdentifyLEDStateOn, want: "ON"},
		{name: "OFF", i: IdentifyLEDStateOff, want: "OFF"},
		{name: "UNKNOWN", i: IdentifyLEDStateUnknown, want: "UNKNOWN"},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.String(); got != tt.want {
				t.Errorf("IdentifyLEDState.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFirmwareMode_String(t *testing.T) {
	tests := []struct {
		name string
		f    FirmwareMode
		want string
	}{
		{name: "LEGACY", f: FirmwareModeLegacy, want: "LEGACY"},
		{name: "UEFI", f: FirmwareModeUEFI, want: "UEFI"},
		{name: "UNKNOWN", f: FirmwareModeUnknown, want: "UNKNOWN"},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.String(); got != tt.want {
				t.Errorf("FirmwareMode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
