package ipmi

// https://www.intel.com/content/dam/www/public/us/en/documents/product-briefs/ipmi-second-gen-interface-spec-v2-rev1-1.pdf

// Appendix H - Sub-function Assignments

type SetSystemBootOptionsFunction = uint8

const (
	ServicePartitionSelector SetSystemBootOptionsFunction = iota + 1
	ServicePartitionScan
	ValidBitClearing
	BootInfoAcknowledge
	BootFlags
	InitiatorInfo
	InitiatorMailbox
	OEMHasHandledBootInfo
	SMSHasHandledBootInfo
	OSServicePartitionHasHandledBootInfo
	OSLoaderHasHandledBootInfo
	BIOSPOSTHasHandledBootInfo
)

type ChassisControlFunction = uint8

const (
	ChassisControlPowerDown ChassisControlFunction = iota
	ChassisControlPowerUp
	ChassisControlPowerCycle
	ChassisControlHardReset
	ChassisControlPulseDiagnosticInterrupt
	ChassisControlInitiateSoftShutdownViaOvertemp
)

type ChassisIdentifyFunction = uint8

const (
	ChassisIdentifyForceOnIndefinitely ChassisIdentifyFunction = iota
)
