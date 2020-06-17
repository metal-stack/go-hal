package ipmi

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBoolValues(t *testing.T) {
	require.Equal(t, uint8(0), False)
	require.Equal(t, uint8(1), True)
}

func Test_fixedBytes(t *testing.T) {
	require.Equal(t, []uint8{116, 101, 115, 116}, fixedBytes("test", 4))
	require.Equal(t, []uint8{116, 101, 115, 116, 0}, fixedBytes("test", 5))
	require.Equal(t, []uint8{116, 101, 115}, fixedBytes("test", 3))
	require.Equal(t, []uint8{}, fixedBytes("test", 0))
}

func TestBitsAndBytes(t *testing.T) {
	n := uint8(4)
	require.False(t, hasBit(n, 7))
	m := setBit(n, 7)
	require.Equal(t, uint8(132), m)
	require.True(t, hasBit(m, 7))
	m = clearBit(m, 7)
	require.False(t, hasBit(m, 7))
	require.Equal(t, n, m)
}
