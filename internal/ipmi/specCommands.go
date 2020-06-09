package ipmi

// https://www.intel.com/content/dam/www/public/us/en/documents/product-briefs/ipmi-second-gen-interface-spec-v2-rev1-1.pdf

// Appendix G - Command Assignments

type GlobalDeviceCommand = uint8

const (
	GetDeviceID GlobalDeviceCommand = iota + 1
	ColdReset
	WarmReset
	GetSelfTestResults
	ManufacturingTestOn
	SetACPIPowerState
	GetACPIPowerState
	GetDeviceGUID
	GetNetFnSupport
	GetCommandSupport
	GetCommandSubFunctionSupport
	GetConfigurableCommands
	GetConfigurableCommandSubFunctions
)

type FirmwareFirewallConfigurationCommand = uint8

const (
	SetCommandEnables FirmwareFirewallConfigurationCommand = iota + 96
	GetCommandEnables
	SetCommandSubFunctionEnables
	GetCommandSubFunctionEnables
	GetOEMNetNnIANASupport
)

type BMCWatchdogTimerCommand = uint8

const (
	ResetWatchdogTimer BMCWatchdogTimerCommand = iota + 34
	_
	SetWatchdogTimer
	GetWatchdogTimer
)

type BMCDeviceAndMessagingCommand = uint8

const (
	SetBMCGlobalEnables BMCDeviceAndMessagingCommand = iota + 46
	GetBMCGlobalEnables
	ClearMessageFlags
	GetMessageFlags
	EnableMessageChannelReceive
	GetMessage
	SendMessage
	ReadEventMessageBuffer
	GetBTInterfaceCapabilities
	GetSystemGUID
	GetChannelAuthenticationCapabilities
	GetSessionChallenge
	ActivateSession
	SetSessionPrivilegeLevel
	CloseSession
	GetSessionInfo
	_
	GetAuthCode
	SetChannelAccess
	GetChannelAccess
	GetChannelInfo
	SetUserAccess
	GetUserAccess
	SetUserName
	GetUserName
	SetUserPassword
	ActivatePayload
	DeactivatePayload
	GetPayloadActivationStatus
	GetPayloadInstanceInfo
	SetUserPayloadAccess
	GetUserPayloadAccess
	GetChannelPayloadSupport
	GetChannelPayloadVersion
	GetChannelOEMPayloadInfo
	_
	MasterReadWrite
	_
	GetChannelCipherSuites
	SuspendResumePayloadEncryption
	SetChannelSecurityKeys
	GetSystemInterfaceCapabilities
	SetSystemInfoParameters
	GetSystemInfoParameters
)

type ChassisDeviceCommand = uint8

const (
	GetChassisCapabilities ChassisDeviceCommand = iota
	GetChassisStatus
	ChassisControl
	ChassisReset
	ChassisIdentify
	SetChassisCapabilities
	SetPowerRestorePolicy
	GetSystemRestartCause
	SetSystemBootOptions
	GetSystemBootOptions
	SetFrontPanelButtonEnables
	SetPowerCycleInterval
	_
	_
	_
	GetPOHCounter
)

type EventCommand = uint8

const (
	SetEventReceiver EventCommand = iota
	GetEventReceiver
	PlatformEvent
)

const (
	GetEventCount EventCommand = iota + 48
	SetEventDestination
	SetEventReceptionState
	SendICMBEventMessage
	GetEventDestination
	GetEventReceptionState
)

type PEFAndAlertingCommand = uint8

const (
	GetPEFCapabilities PEFAndAlertingCommand = iota + 16
	ArmPEFPostponeTimer
	SetPEFConfigurationParameters
	GetPEFConfigurationParameters
	SetLastProcessingEventID
	GetLastProcessingEventID
	AlertImmediate
	PETAcknowledge
)

type SensorDeviceCommand = uint8

const (
	GetDeviceSDRInfo SensorDeviceCommand = iota + 32
	GetDeviceSDR
	ReserveDeviceSDRRepository
	GetSensorReadingFactors
	SetSensorHysteresis
	GetSensorHysteresis
	SetSensorThreshold
	GetSensorThreshold
	SetSensorEventEnable
	GetSensorEventEnable
	ReArmSensorEvents
	GetSensorEventStatus
	_
	GetSensorReading
	SetSensorType
	GetSensorType
	SetSensorReadingAndEventStatus
)

type FRUDeviceCommand = uint8

const (
	GetFRUInventoryAreaInfo FRUDeviceCommand = iota + 16
	ReadFRUData
	WriteFRUData
)

type SDRDeviceCommand = uint8

const (
	GetSDRRepositoryInfo SDRDeviceCommand = iota + 32
	GetSDRRepositoryAllocationInfo
	ReserveSDRRepository
	GetSDR
	AddSDR
	PartialAddSDR
	DeleteSDR
	ClearSDRRepository
	GetSDRRepositoryTime
	SetSDRRepositoryTime
	EnterSDRRepositoryUpdateMode
	ExitSDRRepositoryUpdateMode
	RunInitializationAgent
)

type SELDeviceCommand = uint8

const (
	GetSELInfo SELDeviceCommand = iota + 64
	GetSELAllocationInfo
	ReserveSEL
	GetSELEntry
	AddSELEntry
	PartialAddSELEntry
	DeleteSELEntry
	ClearSEL
	GetSELTime
	SetSELTime
)

const (
	GetAuxiliaryLogStatus SELDeviceCommand = iota + 90
	SetAuxiliaryLogStatus
	GetSELTimeUTCOffset
	SetSELTimeUTCOffset
)

type LANDeviceCommand = uint8

const (
	SetLANConfigurationParameters LANDeviceCommand = iota + 1
	GetLANConfigurationParameters
	SuspendBMCARPs
	GetIPUDPRMCPStatistics
)

type SerialModemDeviceCommand = uint8

const (
	SetSerialModemConfiguration SerialModemDeviceCommand = iota + 16
	GetSerialModemConfiguration
	SetSerialModemMux
	GetTAPResponseCodes
	SetPPPUDPProxyTransmitData
	GetPPPUDPProxyTransmitData
	SetPPPUDPProxyPacket
	GetPPPUDPProxyReceiveData
	SerialModemConnectionActive
	Callback
	SetUserCallbackOptions
	GetUserCallbackOptions
	SetSerialRoutingMux
	_
	_
	_
	SOLActivating
	SetSOLConfigurationParameters
	GetSOLConfigurationParameters
)

type CommandForwardingCommand = uint8

const (
	ForwardedCommand CommandForwardingCommand = iota + 48
	SetForwardedCommands
	GetForwardedCommands
	EnableForwardedCommands
)

type BridgeManagementCommand = uint8

const (
	GetBridgeState BridgeManagementCommand = iota
	SetBridgeState
	GetICMBAddress
	SetICMBAddress
	SetBridgeProxyAddress
	GetBridgeStatistics
	GetICMBCapabilities
	_
	ClearBridgeStatistics
	GetBridgeProxyAddress
	GetICMBConnectorInfo
	GetICMBConnectionID
	SendICMBConnectionID
	ErrorReport BridgeManagementCommand = 255
)

type DiscoveryCommand = uint8

const (
	PrepareForDiscovery DiscoveryCommand = iota + 16
	GetAddresses
	SetDiscovered
	GetChassisDeviceID
	SetChassisDeviceID
)

type BridgingCommand = uint8

const (
	BridgeRequest BridgingCommand = iota + 32
	BridgeMessage
)
