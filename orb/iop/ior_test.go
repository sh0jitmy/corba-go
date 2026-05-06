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
	"testing"

	"github.com/shjtmy/corba-go/orb/cdr"
)

func buildMockIOR() string {
	// Let's build an IIOP profile encapsulation
	profEnc := cdr.NewEncoder(binary.LittleEndian)
	profEnc.EncodeOctet(1) // Minor version
	profEnc.EncodeOctet(2) // Major version (1.2) -> wait, IIOP version is Major then Minor.
	// Actually let's use BigEndian for testing variety
	profEnc = cdr.NewEncoder(binary.LittleEndian)
	profEnc.EncodeOctet(1) // Major version
	profEnc.EncodeOctet(2) // Minor version
	profEnc.EncodeString("127.0.0.1")
	profEnc.EncodeUShort(2809)
	profEnc.EncodeULong(4) // object key length
	profEnc.EncodeOctet('t')
	profEnc.EncodeOctet('e')
	profEnc.EncodeOctet('s')
	profEnc.EncodeOctet('t')

	profData := append([]byte{1}, profEnc.Bytes()...) // 1 indicates LittleEndian for encapsulation

	// Build the main IOR
	iorEnc := cdr.NewEncoder(binary.LittleEndian)
	iorEnc.EncodeString("IDL:test/Object:1.0")
	iorEnc.EncodeULong(1) // 1 profile
	iorEnc.EncodeULong(TAG_INTERNET_IOP)
	iorEnc.EncodeULong(uint32(len(profData))) //#nosec G115 -- test data is always small
	for _, b := range profData {
		iorEnc.EncodeOctet(b)
	}

	iorBytes := append([]byte{1}, iorEnc.Bytes()...) // 1 indicates LittleEndian
	return "IOR:" + hex.EncodeToString(iorBytes)
}

func TestParseIOR(t *testing.T) {
	mockIORStr := buildMockIOR()

	ior, err := ParseIOR(mockIORStr)
	if err != nil {
		t.Fatalf("ParseIOR failed: %v", err)
	}

	if ior.TypeID != "IDL:test/Object:1.0" {
		t.Errorf("Expected TypeID 'IDL:test/Object:1.0', got '%s'", ior.TypeID)
	}

	if len(ior.Profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(ior.Profiles))
	}

	prof := ior.Profiles[0]
	if prof.Tag != TAG_INTERNET_IOP {
		t.Errorf("Expected tag %d, got %d", TAG_INTERNET_IOP, prof.Tag)
	}

	iiopProf, err := ParseIIOPProfile(prof.ProfileData)
	if err != nil {
		t.Fatalf("ParseIIOPProfile failed: %v", err)
	}

	if iiopProf.VersionMajor != 1 || iiopProf.VersionMinor != 2 {
		t.Errorf("Expected IIOP version 1.2, got %d.%d", iiopProf.VersionMajor, iiopProf.VersionMinor)
	}

	if iiopProf.Host != "127.0.0.1" {
		t.Errorf("Expected host '127.0.0.1', got '%s'", iiopProf.Host)
	}

	if iiopProf.Port != 2809 {
		t.Errorf("Expected port 2809, got %d", iiopProf.Port)
	}

	if string(iiopProf.ObjectKey) != "test" {
		t.Errorf("Expected ObjectKey 'test', got '%s'", string(iiopProf.ObjectKey))
	}
}
