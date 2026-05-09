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
	"fmt"
	"log"

	"github.com/sh0jitmy/corba-go/orb/iiop"
	"github.com/sh0jitmy/corba-go/orb/iop"
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

	// Create a stringified IOR using the shared utility
	ior := iop.NewIOR("IDL:Calculator/Math:1.0", "host.docker.internal", 2809, []byte("Test"))
	iorStr := ior.StringifyIOR()

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
