package uuid

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// uuidSize is the size in bytes of a UUID object
const uuidSize = 16

// uid represents a UUID object as defined by RFC 4122.
type uid struct {
	TimeLow          uint32
	TimeMid          uint16
	TimeHiAndVersion uint16
	ClockSeqHiAndRes uint8
	ClockSeqLow      uint8
	Node             [6]byte
}

func Convert(s string) (string, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return "", err
	}
	if !isMixedEncoded(u) {
		return u.String(), nil
	}

	b, err := u.MarshalBinary()
	if err != nil {
		return "", err
	}
	uid, err := fromBytes(b)
	if err != nil {
		return "", err
	}
	buf, err := uid.middleEndianBytes()
	if err != nil {
		return "", err
	}
	uid, err = fromBytes(buf)
	return fmt.Sprintf("%08x-%04x-%04x-%02x%02x-%x", uid.TimeLow, uid.TimeMid, uid.TimeHiAndVersion, uid.ClockSeqHiAndRes, uid.ClockSeqLow, uid.Node), err
}

const tenYears = 10 * 365 * 24 * time.Hour

func isMixedEncoded(u uuid.UUID) bool {
	timeDistance := time.Since(time.Unix(u.Time().UnixTime())).Abs()
	return timeDistance > tenYears
}

// middleEndianBytes returns the UUID encoded in Middle Endian.
func (u uid) middleEndianBytes() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, uuidSize))

	if err := binary.Write(buf, binary.LittleEndian, u.TimeLow); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.LittleEndian, u.TimeMid); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.LittleEndian, u.TimeHiAndVersion); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, u.ClockSeqHiAndRes); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, u.ClockSeqLow); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, u.Node); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UuidFromBytes reads a UUID from a given byte slice.
func fromBytes(buf []byte) (uid, error) {
	var u uid
	reader := bytes.NewReader(buf)

	err := binary.Read(reader, binary.BigEndian, &u)

	return u, err
}
