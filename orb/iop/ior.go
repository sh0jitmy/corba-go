package iop

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/shjtmy/corba-go/orb/cdr"
)

const (
	TAG_INTERNET_IOP uint32 = 0
)

// IOR represents an Interoperable Object Reference.
type IOR struct {
	TypeID   string
	Profiles []TaggedProfile
}

// TaggedProfile represents a generic profile in an IOR.
type TaggedProfile struct {
	Tag         uint32
	ProfileData []byte
}

// IIOPProfile represents the unpacked profile data for TAG_INTERNET_IOP.
type IIOPProfile struct {
	VersionMajor uint8
	VersionMinor uint8
	Host         string
	Port         uint16
	ObjectKey    []byte
}

// ParseIOR parses a stringified IOR (starting with "IOR:").
func ParseIOR(iorStr string) (*IOR, error) {
	if !strings.HasPrefix(iorStr, "IOR:") {
		return nil, errors.New("invalid IOR prefix")
	}

	hexData := iorStr[4:]
	data, err := hex.DecodeString(hexData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode IOR hex: %v", err)
	}

	// Assuming BigEndian for the IOR string encoding by default, unless byte order mark is present.
	// Actually, stringified IOR specifies endianness at the start of CDR encapsulation?
	// Standard CDR encapsulation always starts with a byte order flag (0 = Big Endian, 1 = Little Endian).

	if len(data) < 1 {
		return nil, errors.New("ior data too short")
	}

	byteOrderFlag := data[0]
	var order binary.ByteOrder
	if byteOrderFlag == 0 {
		order = binary.BigEndian
	} else {
		order = binary.LittleEndian
	}

	// Note: The byte order flag might not be at the very start of the IOR structure itself,
	// but IORs are typically marshaled as a CDR stream. Let's assume the stringified IOR is a standard CDR encoded IOR.
	// The standard layout for an encoded IOR does NOT start with a byte order mark directly in the stringified form usually,
	// it starts with the Byte Order of the encapsulation, so data[0] is the byte order.

	dec := cdr.NewDecoder(data[1:], order) // skip byte order flag

	typeID, err := dec.DecodeString()
	if err != nil {
		return nil, fmt.Errorf("failed to decode TypeID: %v", err)
	}

	profileCount, err := dec.DecodeULong()
	if err != nil {
		return nil, fmt.Errorf("failed to decode profile count: %v", err)
	}

	var profiles []TaggedProfile
	for i := uint32(0); i < profileCount; i++ {
		tag, err := dec.DecodeULong()
		if err != nil {
			return nil, fmt.Errorf("failed to decode profile tag: %v", err)
		}

		length, err := dec.DecodeULong()
		if err != nil {
			return nil, fmt.Errorf("failed to decode profile length: %v", err)
		}

		profileData := make([]byte, length)
		for j := uint32(0); j < length; j++ {
			b, err := dec.DecodeOctet()
			if err != nil {
				return nil, fmt.Errorf("failed to decode profile data byte: %v", err)
			}
			profileData[j] = b
		}

		profiles = append(profiles, TaggedProfile{
			Tag:         tag,
			ProfileData: profileData,
		})
	}

	return &IOR{
		TypeID:   typeID,
		Profiles: profiles,
	}, nil
}

// ParseIIOPProfile parses the ProfileData of a TAG_INTERNET_IOP profile.
func ParseIIOPProfile(data []byte) (*IIOPProfile, error) {
	if len(data) < 1 {
		return nil, errors.New("profile data too short")
	}

	// Profile data is an encapsulation. First byte is byte order.
	byteOrderFlag := data[0]
	var order binary.ByteOrder
	if byteOrderFlag == 0 {
		order = binary.BigEndian
	} else {
		order = binary.LittleEndian
	}

	dec := cdr.NewDecoder(data[1:], order)

	major, err := dec.DecodeOctet()
	if err != nil {
		return nil, err
	}

	minor, err := dec.DecodeOctet()
	if err != nil {
		return nil, err
	}

	host, err := dec.DecodeString()
	if err != nil {
		return nil, err
	}

	port, err := dec.DecodeUShort()
	if err != nil {
		return nil, err
	}

	objKeyLen, err := dec.DecodeULong()
	if err != nil {
		return nil, err
	}

	objKey := make([]byte, objKeyLen)
	for i := uint32(0); i < objKeyLen; i++ {
		b, err := dec.DecodeOctet()
		if err != nil {
			return nil, err
		}
		objKey[i] = b
	}

	return &IIOPProfile{
		VersionMajor: major,
		VersionMinor: minor,
		Host:         host,
		Port:         port,
		ObjectKey:    objKey,
	}, nil
}
