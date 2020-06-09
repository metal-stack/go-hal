package ipmi

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSpecSubFunctions(t *testing.T) {
	require.Equal(t, uint8(1), ChassisControlPowerUp)
	require.Equal(t, uint8(2), ChassisControlPowerCycle)
	require.Equal(t, uint8(3), ChassisControlHardReset)
	require.Equal(t, uint8(4), ChassisControlPulseDiagnosticInterrupt)
	require.Equal(t, uint8(5), ChassisControlInitiateSoftShutdownViaOvertemp)
}
