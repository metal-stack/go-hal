package ipmi

// https://www.intel.com/content/dam/www/public/us/en/documents/product-briefs/ipmi-second-gen-interface-spec-v2-rev1-1.pdf

// Appendix H - Sub-function Assignments

type ChassisControlFunction = uint8

const (
	ChassisControlPowerUp ChassisControlFunction = iota + 1
	ChassisControlPowerCycle
	ChassisControlHardReset
	ChassisControlPulseDiagnosticInterrupt
	ChassisControlInitiateSoftShutdownViaOvertemp
)

type ChassisIdentifyFunction = uint8

const (
	ChassisIdentifyForceOnIndefinitely ChassisIdentifyFunction = iota
)
