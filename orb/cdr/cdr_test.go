package cdr

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	enc := NewEncoder(binary.LittleEndian)

	// Encode various types
	enc.EncodeOctet(42)
	enc.EncodeShort(-1234)
	enc.EncodeUShort(5678)
	enc.EncodeLong(-12345678)
	enc.EncodeULong(87654321)
	enc.EncodeLongLong(-123456789012345)
	enc.EncodeULongLong(987654321098765)
	enc.EncodeFloat(3.14159)
	enc.EncodeDouble(2.718281828459)
	enc.EncodeBoolean(true)
	enc.EncodeString("Hello CORBA!")

	data := enc.Bytes()

	dec := NewDecoder(data, binary.LittleEndian)

	// Decode and verify
	vOctet, err := dec.DecodeOctet()
	if err != nil || vOctet != 42 {
		t.Errorf("Octet failed: %v, %v", vOctet, err)
	}

	vShort, err := dec.DecodeShort()
	if err != nil || vShort != -1234 {
		t.Errorf("Short failed: %v, %v", vShort, err)
	}

	vUShort, err := dec.DecodeUShort()
	if err != nil || vUShort != 5678 {
		t.Errorf("UShort failed: %v, %v", vUShort, err)
	}

	vLong, err := dec.DecodeLong()
	if err != nil || vLong != -12345678 {
		t.Errorf("Long failed: %v, %v", vLong, err)
	}

	vULong, err := dec.DecodeULong()
	if err != nil || vULong != 87654321 {
		t.Errorf("ULong failed: %v, %v", vULong, err)
	}

	vLongLong, err := dec.DecodeLongLong()
	if err != nil || vLongLong != -123456789012345 {
		t.Errorf("LongLong failed: %v, %v", vLongLong, err)
	}

	vULongLong, err := dec.DecodeULongLong()
	if err != nil || vULongLong != 987654321098765 {
		t.Errorf("ULongLong failed: %v, %v", vULongLong, err)
	}

	vFloat, err := dec.DecodeFloat()
	if err != nil || math.Abs(float64(vFloat-3.14159)) > 0.00001 {
		t.Errorf("Float failed: %v, %v", vFloat, err)
	}

	vDouble, err := dec.DecodeDouble()
	if err != nil || math.Abs(vDouble-2.718281828459) > 0.00000000001 {
		t.Errorf("Double failed: %v, %v", vDouble, err)
	}

	vBoolean, err := dec.DecodeBoolean()
	if err != nil || vBoolean != true {
		t.Errorf("Boolean failed: %v, %v", vBoolean, err)
	}

	vString, err := dec.DecodeString()
	if err != nil || vString != "Hello CORBA!" {
		t.Errorf("String failed: '%v', %v", vString, err)
	}
}

func TestBaseOffsetAlignment(t *testing.T) {
	enc := NewEncoder(binary.LittleEndian)
	enc.SetBaseOffset(3) // Suppose there are 3 bytes before this buffer
	
	enc.EncodeOctet(1)
	// Current buffer len: 1. Absolute offset: 3 + 1 = 4.
	// Encoding a Long (4 bytes) requires 4-byte alignment.
	// Since absolute offset is 4, it is already aligned! No padding should be added.
	enc.EncodeLong(12345)
	
	data := enc.Bytes()
	if len(data) != 5 { // 1 byte octet + 0 padding + 4 bytes long
		t.Errorf("Expected length 5, got %d", len(data))
	}

	dec := NewDecoder(data, binary.LittleEndian)
	dec.SetBaseOffset(3)
	
	vOctet, _ := dec.DecodeOctet()
	if vOctet != 1 {
		t.Errorf("Expected octet 1, got %d", vOctet)
	}
	
	vLong, err := dec.DecodeLong()
	if err != nil || vLong != 12345 {
		t.Errorf("Long failed: %v, %v", vLong, err)
	}
}

