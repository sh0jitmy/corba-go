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

package giop

import (
	"encoding/binary"
	"errors"

	"github.com/shjtmy/corba-go/orb/cdr"
)

// MsgType represents the GIOP message type.
type MsgType uint8

const (
	RequestMsg       MsgType = 0
	ReplyMsg         MsgType = 1
	CancelRequestMsg MsgType = 2
	LocateRequestMsg MsgType = 3
	LocateReplyMsg   MsgType = 4
	CloseConnection  MsgType = 5
	MessageError     MsgType = 6
	Fragment         MsgType = 7
)

// Header is the standard 12-byte GIOP header (GIOP 1.0, 1.1, 1.2).
type Header struct {
	Magic       [4]byte
	Version     Version
	Flags       uint8
	MessageType MsgType
	MessageSize uint32
}

// Version represents the GIOP version.
type Version struct {
	Major uint8
	Minor uint8
}

// RequestHeader represents a GIOP 1.2 Request Header.
type RequestHeader struct {
	RequestID       uint32
	ResponseFlags   uint8
	Reserved        [3]byte
	TargetAddress   []byte // Simplify TargetAddress handling for now (ObjectKey)
	Operation       string
	ServiceContexts []ServiceContext
	// Note: Requesting Principal is deprecated in GIOP 1.2, ignored here.
}

// ServiceContext represents a service context element.
type ServiceContext struct {
	ContextID   uint32
	ContextData []byte
}

// ReadHeader reads a GIOP header from a buffer.
func ReadHeader(data []byte) (*Header, error) {
	if len(data) < 12 {
		return nil, errors.New("giop header too short")
	}

	h := &Header{}
	copy(h.Magic[:], data[0:4])
	if string(h.Magic[:]) != "GIOP" {
		return nil, errors.New("invalid magic number, expected GIOP")
	}

	h.Version.Major = data[4]
	h.Version.Minor = data[5]
	h.Flags = data[6]
	h.MessageType = MsgType(data[7])

	var order binary.ByteOrder
	if h.Flags&0x01 == 1 {
		order = binary.LittleEndian
	} else {
		order = binary.BigEndian
	}

	h.MessageSize = order.Uint32(data[8:12])

	return h, nil
}

// WriteHeader writes a GIOP header to an encoder.
func WriteHeader(enc *cdr.Encoder, h *Header) {
	enc.EncodeOctet(h.Magic[0])
	enc.EncodeOctet(h.Magic[1])
	enc.EncodeOctet(h.Magic[2])
	enc.EncodeOctet(h.Magic[3])
	enc.EncodeOctet(h.Version.Major)
	enc.EncodeOctet(h.Version.Minor)
	enc.EncodeOctet(h.Flags)
	enc.EncodeOctet(uint8(h.MessageType))
	enc.EncodeULong(h.MessageSize)
}

// EncodeRequest creates a full GIOP request message.
func EncodeRequest(reqID uint32, operation string, targetObjKey []byte, args []byte) ([]byte, error) {
	// GIOP 1.2 assumes 8-byte alignment after header.
	// We'll use LittleEndian for encoding by default (Flag=1).
	enc := cdr.NewEncoder(binary.LittleEndian)

	// --- Payload Encoding ---
	// Encode Request ID
	enc.EncodeULong(reqID)
	// Encode Response Flags (0x03 means expecting a reply)
	enc.EncodeOctet(3)
	// Reserved
	enc.EncodeOctet(0)
	enc.EncodeOctet(0)
	enc.EncodeOctet(0)

	// TargetAddress (Disposition: 0 = KeyAddr)
	enc.EncodeShort(0)
	// Object Key length and data
	enc.EncodeULong(uint32(len(targetObjKey))) //#nosec G115 -- object key length will not exceed uint32 max
	for _, b := range targetObjKey {
		enc.EncodeOctet(b)
	}

	// Operation name
	enc.EncodeString(operation)

	// Service Context list (empty for now)
	enc.EncodeULong(0)

	// For GIOP 1.2 Request, the request body is 8-byte aligned relative to the start of the GIOP message.
	// The GIOP header is 12 bytes. We need to align the start of arguments.
	// Since cdr.Encoder doesn't know the absolute start, we simulate it.
	// In our encoder, offset is len(buffer). Absolute = len(buffer) + 12.
	// We want (absolute) % 8 == 0.

	currentLen := len(enc.Bytes())
	absPos := currentLen + 12
	padding := (8 - (absPos % 8)) % 8
	for i := 0; i < padding; i++ {
		enc.EncodeOctet(0)
	}

	// Append arguments
	payload := enc.Bytes()
	payload = append(payload, args...)

	// --- Header Encoding ---
	headerEnc := cdr.NewEncoder(binary.LittleEndian)
	h := &Header{
		Magic:       [4]byte{'G', 'I', 'O', 'P'},
		Version:     Version{Major: 1, Minor: 2},
		Flags:       1, // Little Endian
		MessageType: RequestMsg,
		MessageSize: uint32(len(payload)), //#nosec G115 -- payload length is bounded
	}
	WriteHeader(headerEnc, h)

	// Combine Header + Payload
	msg := append(headerEnc.Bytes(), payload...)
	return msg, nil
}
