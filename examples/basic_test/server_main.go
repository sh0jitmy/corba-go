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

package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/shjtmy/corba-go/orb/cdr"
	"github.com/shjtmy/corba-go/orb/iiop"
	"github.com/shjtmy/corba-go/orb/iop"
)

type MathImpl struct{}

func (m *MathImpl) Add(a int32, b int32) (int32, error) {
	fmt.Printf("Received Add(%d, %d)\n", a, b)
	return a + b, nil
}

func (m *MathImpl) Echo(msg string) (string, error) {
	fmt.Printf("Received Echo(%s)\n", msg)
	return "Echo from Go: " + msg, nil
}

func main() {
	server, err := iiop.NewServer(2809) // Default port for demo
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	impl := &MathImpl{}
	handler := Calculator_Math_Skeleton(impl)

	// Create a dummy stringified IOR
	profEnc := cdr.NewEncoder(binary.LittleEndian)
	profEnc.EncodeOctet(1) // byte order (little endian)
	profEnc.EncodeOctet(1) // Major (1)
	profEnc.EncodeOctet(2) // Minor (2)
	profEnc.EncodeString("host.docker.internal")
	profEnc.EncodeUShort(2809)
	profEnc.EncodeULong(4) // object key length
	profEnc.EncodeOctet('T')
	profEnc.EncodeOctet('e')
	profEnc.EncodeOctet('s')
	profEnc.EncodeOctet('t')
	profEnc.EncodeULong(0) // sequence of TaggedComponent length (0 components)

	profData := profEnc.Bytes()

	iorEnc := cdr.NewEncoder(binary.LittleEndian)
	iorEnc.EncodeOctet(1) // byte order (little endian)
	iorEnc.EncodeString("IDL:Calculator/Math:1.0")
	iorEnc.EncodeULong(1) // 1 profile
	iorEnc.EncodeULong(iop.TAG_INTERNET_IOP)
	iorEnc.EncodeULong(uint32(len(profData))) //#nosec G115 -- profData length is always small
	for _, b := range profData {
		iorEnc.EncodeOctet(b)
	}

	iorBytes := iorEnc.Bytes()
	iorStr := "IOR:" + hex.EncodeToString(iorBytes)

	fmt.Println("Server started on :2809")
	fmt.Println("IOR:")
	fmt.Println(iorStr)

	err = server.Serve(func(objectKey []byte, operation string, reqPayload []byte) ([]byte, error) {
		if string(objectKey) != "Test" {
			return nil, fmt.Errorf("unknown object key: %s", string(objectKey))
		}
		return handler(objectKey, operation, reqPayload)
	})
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
