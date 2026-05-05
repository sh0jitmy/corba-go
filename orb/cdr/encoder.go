package cdr

import (
	"encoding/binary"
	"math"
)

// Encoder handles the marshaling of Go types into CORBA CDR format.
type Encoder struct {
	buffer     []byte
	order      binary.ByteOrder
	baseOffset int
}

// NewEncoder creates a new CDR encoder.
// CORBA commonly defaults to LittleEndian for modern architectures, but it can be specified.
func NewEncoder(order binary.ByteOrder) *Encoder {
	if order == nil {
		order = binary.LittleEndian
	}
	return &Encoder{
		buffer: make([]byte, 0, 128),
		order:  order,
	}
}

// SetBaseOffset sets the absolute base offset (e.g. 12 for GIOP header) used for alignment calculation.
func (e *Encoder) SetBaseOffset(offset int) {
	e.baseOffset = offset
}

// Bytes returns the encoded CDR byte slice.
func (e *Encoder) Bytes() []byte {
	return e.buffer
}

// align padding to a given boundary (e.g., 2, 4, 8)
func (e *Encoder) align(boundary int) {
	absOffset := len(e.buffer) + e.baseOffset
	padding := (boundary - (absOffset % boundary)) % boundary
	for i := 0; i < padding; i++ {
		e.buffer = append(e.buffer, 0)
	}
}

// EncodeOctet encodes a single byte (uint8). No alignment required.
func (e *Encoder) EncodeOctet(v uint8) {
	e.buffer = append(e.buffer, v)
}

// EncodeShort encodes a 16-bit integer (int16). Aligned to 2 bytes.
func (e *Encoder) EncodeShort(v int16) {
	e.align(2)
	b := make([]byte, 2)
	e.order.PutUint16(b, uint16(v))
	e.buffer = append(e.buffer, b...)
}

// EncodeUShort encodes an unsigned 16-bit integer (uint16). Aligned to 2 bytes.
func (e *Encoder) EncodeUShort(v uint16) {
	e.align(2)
	b := make([]byte, 2)
	e.order.PutUint16(b, v)
	e.buffer = append(e.buffer, b...)
}

// EncodeLong encodes a 32-bit integer (int32). Aligned to 4 bytes.
func (e *Encoder) EncodeLong(v int32) {
	e.align(4)
	b := make([]byte, 4)
	e.order.PutUint32(b, uint32(v))
	e.buffer = append(e.buffer, b...)
}

// EncodeULong encodes an unsigned 32-bit integer (uint32). Aligned to 4 bytes.
func (e *Encoder) EncodeULong(v uint32) {
	e.align(4)
	b := make([]byte, 4)
	e.order.PutUint32(b, v)
	e.buffer = append(e.buffer, b...)
}

// EncodeLongLong encodes a 64-bit integer (int64). Aligned to 8 bytes.
func (e *Encoder) EncodeLongLong(v int64) {
	e.align(8)
	b := make([]byte, 8)
	e.order.PutUint64(b, uint64(v))
	e.buffer = append(e.buffer, b...)
}

// EncodeULongLong encodes an unsigned 64-bit integer (uint64). Aligned to 8 bytes.
func (e *Encoder) EncodeULongLong(v uint64) {
	e.align(8)
	b := make([]byte, 8)
	e.order.PutUint64(b, v)
	e.buffer = append(e.buffer, b...)
}

// EncodeFloat encodes a 32-bit float. Aligned to 4 bytes.
func (e *Encoder) EncodeFloat(v float32) {
	e.EncodeULong(math.Float32bits(v))
}

// EncodeDouble encodes a 64-bit float. Aligned to 8 bytes.
func (e *Encoder) EncodeDouble(v float64) {
	e.EncodeULongLong(math.Float64bits(v))
}

// EncodeBoolean encodes a boolean. Encoded as an octet (0 or 1).
func (e *Encoder) EncodeBoolean(v bool) {
	if v {
		e.EncodeOctet(1)
	} else {
		e.EncodeOctet(0)
	}
}

// EncodeString encodes a string.
// A string is encoded as an unsigned long indicating its length (including the null terminator),
// followed by the string characters and a null terminator.
func (e *Encoder) EncodeString(v string) {
	// Length includes the null terminator
	length := uint32(len(v) + 1)
	e.EncodeULong(length)
	
	e.buffer = append(e.buffer, []byte(v)...)
	e.buffer = append(e.buffer, 0) // null terminator
}
