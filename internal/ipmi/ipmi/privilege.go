package ipmi

// Privilege of an IPMI user
type Privilege = uint8

const (
	// Callback IPMI privilege
	CallbackPrivilege Privilege = iota + 1
	// User IPMI privilege
	UserPrivilege
	// Operator IPMI privilege
	OperatorPrivilege
	// Administrator IPMI privilege
	AdministratorPrivilege
	// OEM IPMI privilege
	OEMPrivilege
	// NoAccess IPMI privilege
	NoAccessPrivilege
)
