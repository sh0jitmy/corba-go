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
	"os"

	"github.com/sh0jitmy/corba-go/orb/iiop"
	"github.com/sh0jitmy/corba-go/orb/iop"
	"github.com/sh0jitmy/corba-go/services/naming"
)

func main() {
	// Connect to the Naming Service
	client, err := iiop.NewClient("localhost", 2809)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = client.Close() }()

	namingContext := naming.NewCosNaming_NamingContext_Stub(client, []byte("NameService"))

	name := naming.CosNaming_Name{
		&naming.CosNaming_NameComponent{Id: naming.CosNaming_Istring("Math"), Kind: naming.CosNaming_Istring("Service")},
	}

	// Resolve the name to get an IOR string
	iorStr, err := namingContext.Resolve(name)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Resolved IOR: %s\n", iorStr)

	// Parse the IOR string to extract host, port, and object key
	ior, err := iop.ParseIOR(iorStr)
	if err != nil {
		panic(err)
	}

	if len(ior.Profiles) == 0 {
		panic("Error: no profiles in IOR")
	}

	prof, err := iop.ParseIIOPProfile(ior.Profiles[0].ProfileData)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Connecting to %s:%d, objectKey=%s\n", prof.Host, prof.Port, string(prof.ObjectKey))

	// Connect to the target server using the resolved IOR
	mathClient, err := iiop.NewClient(prof.Host, prof.Port)
	if err != nil {
		panic(err)
	}
	defer func() { _ = mathClient.Close() }()

	calculator := NewCalculator_Math_Stub(mathClient, prof.ObjectKey)

	ret, err := calculator.Add(int32(1), int32(2))
	if err != nil {
		panic(err)
	}
	fmt.Printf("1 + 2 = %d\n", ret)
}
