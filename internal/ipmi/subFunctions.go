package ipmi

// https://www.intel.com/content/dam/www/public/us/en/documents/product-briefs/ipmi-second-gen-interface-spec-v2-rev1-1.pdf

// Appendix H - Sub-function Assignments

// ChassisControlFunction is the action requested by the Chassis Control command (section 28.3).
type ChassisControlFunction = uint8

const (
	ChassisControlPowerDown ChassisControlFunction = iota
	ChassisControlPowerUp
	ChassisControlPowerCycle
	ChassisControlHardReset
	ChassisControlPulseDiagnosticInterrupt
	ChassisControlInitiateSoftShutdownViaOvertemp
)
