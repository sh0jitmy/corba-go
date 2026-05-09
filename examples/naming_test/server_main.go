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
	"os"

	"github.com/sh0jitmy/corba-go/orb/iiop"
	"github.com/sh0jitmy/corba-go/orb/iop"
	"github.com/sh0jitmy/corba-go/services/naming"
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
	portNumber := uint16(42809)
	server, err := iiop.NewServer(portNumber)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	impl := &MathImpl{}
	handler := Calculator_Math_Skeleton(impl)

	// Create a stringified IOR using the shared utility
	ior := iop.NewIOR("IDL:Calculator/Math:1.0", "localhost", portNumber, []byte("test"))
	iorStr := ior.StringifyIOR()

	fmt.Printf("Server started on :%d\n", portNumber)
	fmt.Println("IOR:")
	fmt.Println(iorStr)

	//naming service
	client, err := iiop.NewClient("localhost", 2809)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	namingContext := naming.NewCosNaming_NamingContext_Stub(client, []byte("NameService"))

	name := naming.CosNaming_Name{
		&naming.CosNaming_NameComponent{Id: naming.CosNaming_Istring("Math"), Kind: naming.CosNaming_Istring("Service")},
	}

	err = namingContext.Bind(name, iorStr)
	if err != nil {
		fmt.Printf("Error binding name: %v\n", err)
		os.Exit(1)
	}

	err = server.Serve(func(objectKey []byte, operation string, reqPayload []byte) ([]byte, error) {
		fmt.Printf("Received request: %s\n", string(objectKey))
		if string(objectKey) != "test" {
			return nil, fmt.Errorf("unknown object key: %s", string(objectKey))
		}
		return handler(objectKey, operation, reqPayload)
	})
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
