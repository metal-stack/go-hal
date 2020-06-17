package ipmi

type Bool = uint8

const (
	False Bool = iota
	True
)

func fixedBytes(s string, length int) []uint8 {
	bb := make([]byte, length)
	for i, b := range []byte(s) {
		if i == length {
			break
		}
		bb[i] = b
	}
	if len(bb) > length {
		bb = bb[:length]
	}
	return bb
}

func setBit(n uint8, pos int) uint8 {
	return n | 1<<pos
}

func clearBit(n uint8, pos int) uint8 {
	return n &^ (1 << pos)
}

func hasBit(n uint8, pos int) bool {
	return n&(1<<pos) > 0
}
