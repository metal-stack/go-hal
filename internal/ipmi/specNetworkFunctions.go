package ipmi

// https://www.intel.com/content/dam/www/public/us/en/documents/product-briefs/ipmi-second-gen-interface-spec-v2-rev1-1.pdf

// 5.1 Network Function Codes

type NetworkFunction = uint8

const (
	ChassisNetworkFunction NetworkFunction = iota
	ChassisResponse
	BridgeNetworkFunction
	BridgeResponse
	SensorEventNetworkFunction
	SensorEventResponse
	AppNetworkFunction
	AppResponse
	FirmwareNetworkFunction
	FirmwareResponse
	StorageNetworkFunction
	StorageResponse
	TransportNetworkFunction
	TransportResponse
)
