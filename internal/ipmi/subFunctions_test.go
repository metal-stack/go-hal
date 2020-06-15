package ipmi

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSpecSubFunctions(t *testing.T) {
	require.Equal(t, uint8(1), ServicePartitionSelector)
	require.Equal(t, uint8(2), ServicePartitionScan)
	require.Equal(t, uint8(3), ValidBitClearing)
	require.Equal(t, uint8(4), BootInfoAcknowledge)
	require.Equal(t, uint8(5), BootFlags)
	require.Equal(t, uint8(6), InitiatorInfo)
	require.Equal(t, uint8(7), InitiatorMailbox)
	require.Equal(t, uint8(8), OEMHasHandledBootInfo)
	require.Equal(t, uint8(9), SMSHasHandledBootInfo)
	require.Equal(t, uint8(10), OSServicePartitionHasHandledBootInfo)
	require.Equal(t, uint8(11), OSLoaderHasHandledBootInfo)
	require.Equal(t, uint8(12), BIOSPOSTHasHandledBootInfo)

	require.Equal(t, uint8(1), ChassisControlPowerUp)
	require.Equal(t, uint8(2), ChassisControlPowerCycle)
	require.Equal(t, uint8(3), ChassisControlHardReset)
	require.Equal(t, uint8(4), ChassisControlPulseDiagnosticInterrupt)
	require.Equal(t, uint8(5), ChassisControlInitiateSoftShutdownViaOvertemp)

	require.Equal(t, uint8(0), ChassisIdentifyForceOnIndefinitely)
}
