// Copyright 2026- The corba-go Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iop

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/sh0jitmy/corba-go/orb/cdr"
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

// NewIIOPProfile creates a new IIOPProfile with the given parameters.
func NewIIOPProfile(host string, port uint16, objectKey []byte) *IIOPProfile {
	return &IIOPProfile{
		VersionMajor: 1,
		VersionMinor: 2,
		Host:         host,
		Port:         port,
		ObjectKey:    objectKey,
	}
}

// Encode serializes the IIOP profile into a CDR encapsulation byte slice
// (including the leading byte-order flag).
func (p *IIOPProfile) Encode() []byte {
	enc := cdr.NewEncoder(binary.LittleEndian)
	enc.EncodeOctet(p.VersionMajor)
	enc.EncodeOctet(p.VersionMinor)
	enc.EncodeString(p.Host)
	enc.EncodeUShort(p.Port)
	enc.EncodeULong(uint32(len(p.ObjectKey))) //#nosec G115 -- object key length is always small
	for _, b := range p.ObjectKey {
		enc.EncodeOctet(b)
	}
	enc.EncodeULong(0) // 0 TaggedComponents

	// Wrap as CDR encapsulation: byte-order flag (1 = LittleEndian) + encoded data
	return append([]byte{1}, enc.Bytes()...)
}

// NewIOR creates a new IOR with a single IIOP profile.
func NewIOR(typeID string, host string, port uint16, objectKey []byte) *IOR {
	prof := NewIIOPProfile(host, port, objectKey)
	return &IOR{
		TypeID: typeID,
		Profiles: []TaggedProfile{
			{
				Tag:         TAG_INTERNET_IOP,
				ProfileData: prof.Encode(),
			},
		},
	}
}

// StringifyIOR encodes the IOR into the standard "IOR:..." stringified format.
func (ior *IOR) StringifyIOR() string {
	enc := cdr.NewEncoder(binary.LittleEndian)
	enc.EncodeString(ior.TypeID)
	enc.EncodeULong(uint32(len(ior.Profiles))) //#nosec G115 -- profile count is always small
	for _, p := range ior.Profiles {
		enc.EncodeULong(p.Tag)
		enc.EncodeULong(uint32(len(p.ProfileData))) //#nosec G115 -- profile data is bounded
		for _, b := range p.ProfileData {
			enc.EncodeOctet(b)
		}
	}

	// Wrap as CDR encapsulation: byte-order flag (1 = LittleEndian) + encoded data
	iorBytes := append([]byte{1}, enc.Bytes()...)
	return "IOR:" + hex.EncodeToString(iorBytes)
}
