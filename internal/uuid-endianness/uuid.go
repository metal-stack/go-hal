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

// Redfish API and probably the DHCP Request return the GUID of a machine encoded either
// in middleEndianBytes
// or as normal string
//
// To detect which one is actually in use, we extract the creation time of the uuid,
// if this time is somehow reasonable, we return the string of the GUID as we got it,
// we need to convert it to mixed endian
// https://en.wikipedia.org/wiki/Universally_unique_identifier#Encoding
func Convert(s string) (string, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return "", err
	}
	if isNotEncoded(u) {
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

// thirtyYears is an educated guess for a plausible time stored in the uuid.
// RFC states that the time is stored in 100s of nanos since 15 Okt 1582 as defined in RFC 4122
// We check if this time is not more than 30 years apart from now.
// If the uid returned from the BMC is mixedEndian encoded, the time extracted is usually in the year 4000 or so.
const thirtyYears = 30 * 365 * 24 * time.Hour

func isNotEncoded(u uuid.UUID) bool {
	timeDistance := time.Since(time.Unix(u.Time().UnixTime())).Abs()
	return timeDistance < thirtyYears
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
