package uuid

import (
	"bytes"
	"encoding/binary"
	"fmt"

	guuid "github.com/google/uuid"
)

// UuidSize is the size in bytes of a UUID object
const UuidSize = 16

// Uuid represents a UUID object as defined by RFC 4122.
type Uuid struct {
	TimeLow          uint32
	TimeMid          uint16
	TimeHiAndVersion uint16
	ClockSeqHiAndRes uint8
	ClockSeqLow      uint8
	Node             [6]byte
}

// String implements the fmt.Stringer interface
func (u Uuid) String() string {
	return fmt.Sprintf("%x-%x-%x-%x%x-%x", u.TimeLow, u.TimeMid, u.TimeHiAndVersion, u.ClockSeqHiAndRes, u.ClockSeqLow, u.Node)
}

func FromString(s string) (Uuid, error) {
	var uid Uuid
	u, err := guuid.Parse(s)
	if err != nil {
		return uid, err
	}
	b, err := u.MarshalBinary()
	if err != nil {
		return uid, err
	}
	return FromBytes(b)
}

// ToMiddleEndian encodes the UUID into a middle-endian UUID.
//
// A middle-endian encoded UUID represents a UUID where the first
// three groups are Little Endian encoded, while the rest of the
// groups are Big Endian encoded.
func (u Uuid) ToMiddleEndian() (Uuid, error) {
	buf, err := u.MiddleEndianBytes()
	if err != nil {
		return Uuid{}, err
	}

	return FromBytes(buf)
}

// BigEndianBytes returns the UUID encoded in Big Endian.
func (u Uuid) BigEndianBytes() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, UuidSize))

	err := binary.Write(buf, binary.BigEndian, u)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// LittleEndianBytes returns the UUID encoded in Little Endian.
func (u Uuid) LittleEndianBytes() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, UuidSize))

	err := binary.Write(buf, binary.LittleEndian, u)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// MiddleEndianBytes returns the UUID encoded in Middle Endian.
func (u Uuid) MiddleEndianBytes() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, UuidSize))

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
func FromBytes(buf []byte) (Uuid, error) {
	var u Uuid
	reader := bytes.NewReader(buf)

	err := binary.Read(reader, binary.BigEndian, &u)

	return u, err
}
