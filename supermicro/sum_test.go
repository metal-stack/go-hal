package supermicro

import (
	"testing"
)

func TestParseUUIDLine(t *testing.T) {
	l := "UUID                   {SYUU}    = 00000000-0000-0000-0000-AC1F6B7AEB76 // 4-2-2-2-6 formatted 16-byte hex values"
	e := "00000000-0000-0000-0000-ac1f6b7aeb76"
	a := parseUUIDLine(l)
	if e != a {
		t.Fail()
	}
}
