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

package cdr

import (
	"fmt"
)

// TCKind represents the CORBA TypeCode kind.
type TCKind uint32

const (
	tk_null    TCKind = 0
	tk_void    TCKind = 1
	tk_short   TCKind = 2
	tk_long    TCKind = 3
	tk_ushort  TCKind = 4
	tk_ulong   TCKind = 5
	tk_float   TCKind = 6
	tk_double  TCKind = 7
	tk_boolean TCKind = 8
	tk_char    TCKind = 9
	tk_octet   TCKind = 10
	tk_any     TCKind = 11
	tk_string  TCKind = 18
)

// Any represents the CORBA Any type.
type Any struct {
	Type  TCKind
	Value interface{}
}

// EncodeAny encodes an Any type into the encoder.
func (e *Encoder) EncodeAny(a Any) {
	// Encode TypeCode (For simple types, TypeCode is just the TCKind as ulong)
	e.EncodeULong(uint32(a.Type))

	// Some TypeCodes require additional parameters (like string length max),
	// but for simplicity we assume simple TypeCodes with no parameters (length=0 for string)
	if a.Type == tk_string {
		e.EncodeULong(0) // max length (0 = unbounded)
	}

	// Encode Value
	switch a.Type {
	case tk_long:
		e.EncodeLong(a.Value.(int32))
	case tk_octet:
		e.EncodeOctet(a.Value.(byte))
	case tk_string:
		e.EncodeString(a.Value.(string))
	default:
		// Unsupported kind for this simple implementation
		panic(fmt.Sprintf("EncodeAny: unsupported TCKind %d", a.Type))
	}
}

// DecodeAny decodes an Any type from the decoder.
func (d *Decoder) DecodeAny() (Any, error) {
	var a Any
	kind, err := d.DecodeULong()
	if err != nil {
		return a, err
	}
	a.Type = TCKind(kind)

	if a.Type == tk_string {
		// skip max length
		_, _ = d.DecodeULong()
	}

	switch a.Type {
	case tk_long:
		val, err := d.DecodeLong()
		a.Value = val
		return a, err
	case tk_octet:
		val, err := d.DecodeOctet()
		a.Value = val
		return a, err
	case tk_string:
		val, err := d.DecodeString()
		a.Value = val
		return a, err
	default:
		return a, fmt.Errorf("DecodeAny: unsupported TCKind %d", a.Type)
	}
}
