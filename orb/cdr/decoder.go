package cdr

import (
	"encoding/binary"
	"errors"
	"math"
)

// Decoder handles the unmarshaling of CORBA CDR format into Go types.
type Decoder struct {
	buffer     []byte
	offset     int
	order      binary.ByteOrder
	baseOffset int
}

// NewDecoder creates a new CDR decoder from a byte slice.
func NewDecoder(data []byte, order binary.ByteOrder) *Decoder {
	if order == nil {
		order = binary.LittleEndian
	}
	return &Decoder{
		buffer: data,
		offset: 0,
		order:  order,
	}
}

// SetBaseOffset sets the absolute base offset used for alignment.
func (d *Decoder) SetBaseOffset(offset int) {
	d.baseOffset = offset
}

// align advances the offset to meet the required boundary.
func (d *Decoder) align(boundary int) error {
	absOffset := d.offset + d.baseOffset
	padding := (boundary - (absOffset % boundary)) % boundary
	if d.offset+padding > len(d.buffer) {
		return errors.New("cdr decoder: out of bounds during alignment")
	}
	d.offset += padding
	return nil
}

// DecodeOctet decodes a single byte (uint8).
func (d *Decoder) DecodeOctet() (uint8, error) {
	if d.offset+1 > len(d.buffer) {
		return 0, errors.New("cdr decoder: out of bounds")
	}
	v := d.buffer[d.offset]
	d.offset += 1
	return v, nil
}

// DecodeShort decodes a 16-bit integer (int16).
func (d *Decoder) DecodeShort() (int16, error) {
	if err := d.align(2); err != nil {
		return 0, err
	}
	if d.offset+2 > len(d.buffer) {
		return 0, errors.New("cdr decoder: out of bounds")
	}
	v := d.order.Uint16(d.buffer[d.offset : d.offset+2])
	d.offset += 2
	return int16(v), nil
}

// DecodeUShort decodes an unsigned 16-bit integer (uint16).
func (d *Decoder) DecodeUShort() (uint16, error) {
	if err := d.align(2); err != nil {
		return 0, err
	}
	if d.offset+2 > len(d.buffer) {
		return 0, errors.New("cdr decoder: out of bounds")
	}
	v := d.order.Uint16(d.buffer[d.offset : d.offset+2])
	d.offset += 2
	return v, nil
}

// DecodeLong decodes a 32-bit integer (int32).
func (d *Decoder) DecodeLong() (int32, error) {
	if err := d.align(4); err != nil {
		return 0, err
	}
	if d.offset+4 > len(d.buffer) {
		return 0, errors.New("cdr decoder: out of bounds")
	}
	v := d.order.Uint32(d.buffer[d.offset : d.offset+4])
	d.offset += 4
	return int32(v), nil
}

// DecodeULong decodes an unsigned 32-bit integer (uint32).
func (d *Decoder) DecodeULong() (uint32, error) {
	if err := d.align(4); err != nil {
		return 0, err
	}
	if d.offset+4 > len(d.buffer) {
		return 0, errors.New("cdr decoder: out of bounds")
	}
	v := d.order.Uint32(d.buffer[d.offset : d.offset+4])
	d.offset += 4
	return v, nil
}

// DecodeLongLong decodes a 64-bit integer (int64).
func (d *Decoder) DecodeLongLong() (int64, error) {
	if err := d.align(8); err != nil {
		return 0, err
	}
	if d.offset+8 > len(d.buffer) {
		return 0, errors.New("cdr decoder: out of bounds")
	}
	v := d.order.Uint64(d.buffer[d.offset : d.offset+8])
	d.offset += 8
	return int64(v), nil
}

// DecodeULongLong decodes an unsigned 64-bit integer (uint64).
func (d *Decoder) DecodeULongLong() (uint64, error) {
	if err := d.align(8); err != nil {
		return 0, err
	}
	if d.offset+8 > len(d.buffer) {
		return 0, errors.New("cdr decoder: out of bounds")
	}
	v := d.order.Uint64(d.buffer[d.offset : d.offset+8])
	d.offset += 8
	return v, nil
}

// DecodeFloat decodes a 32-bit float.
func (d *Decoder) DecodeFloat() (float32, error) {
	v, err := d.DecodeULong()
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(v), nil
}

// DecodeDouble decodes a 64-bit float.
func (d *Decoder) DecodeDouble() (float64, error) {
	v, err := d.DecodeULongLong()
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(v), nil
}

// DecodeBoolean decodes a boolean.
func (d *Decoder) DecodeBoolean() (bool, error) {
	v, err := d.DecodeOctet()
	if err != nil {
		return false, err
	}
	return v != 0, nil
}

// DecodeString decodes a string.
func (d *Decoder) DecodeString() (string, error) {
	length, err := d.DecodeULong()
	if err != nil {
		return "", err
	}
	if length == 0 {
		return "", nil // should at least have null terminator, but handle gracefully
	}
	
	if uint32(d.offset)+length > uint32(len(d.buffer)) {
		return "", errors.New("cdr decoder: out of bounds for string")
	}
	
	// length includes the null terminator
	strData := d.buffer[d.offset : uint32(d.offset)+length-1]
	d.offset += int(length)
	
	return string(strData), nil
}

// Rest returns the unread portion of the buffer.
func (d *Decoder) Rest() []byte {
	if d.offset >= len(d.buffer) {
		return nil
	}
	return d.buffer[d.offset:]
}

