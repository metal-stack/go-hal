package ipmi

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSpecNetFunctions(t *testing.T) {
	require.Equal(t, uint8(0x00), ChassisNetworkFunction)
	require.Equal(t, uint8(0x01), ChassisResponse)
	require.Equal(t, uint8(0x02), BridgeNetworkFunction)
	require.Equal(t, uint8(0x03), BridgeResponse)
	require.Equal(t, uint8(0x04), SensorEventNetworkFunction)
	require.Equal(t, uint8(0x05), SensorEventResponse)
	require.Equal(t, uint8(0x06), AppNetworkFunction)
	require.Equal(t, uint8(0x07), AppResponse)
	require.Equal(t, uint8(0x08), FirmwareNetworkFunction)
	require.Equal(t, uint8(0x09), FirmwareResponse)
	require.Equal(t, uint8(0x0A), StorageNetworkFunction)
	require.Equal(t, uint8(0x0B), StorageResponse)
	require.Equal(t, uint8(0x0C), TransportNetworkFunction)
	require.Equal(t, uint8(0x0D), TransportResponse)
}
