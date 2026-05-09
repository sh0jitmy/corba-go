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

// supplier_main.go demonstrates a CosEvent push supplier that connects to
// the EventService, obtains a ProxyPushConsumer, and pushes events.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sh0jitmy/corba-go/orb/cdr"
	"github.com/sh0jitmy/corba-go/orb/iiop"
	"github.com/sh0jitmy/corba-go/services/event"
)

func main() {
	// Connect to the Event Service
	client, err := iiop.NewClient("localhost", 2810)
	if err != nil {
		fmt.Printf("Error connecting to EventService: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = client.Close() }()

	// 1. Get EventChannel
	eventChannel := event.NewCosEvent_EventChannel_Stub(client, []byte("EventChannel"))

	// 2. Get SupplierAdmin from EventChannel
	supplierAdminKey, err := eventChannel.For_suppliers()
	if err != nil {
		fmt.Printf("Error getting SupplierAdmin: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Got SupplierAdmin: %s\n", supplierAdminKey)

	supplierAdmin := event.NewCosEvent_SupplierAdmin_Stub(client, []byte(supplierAdminKey))

	// 3. Obtain a ProxyPushConsumer from SupplierAdmin
	proxyPushConsumerKey, err := supplierAdmin.Obtain_push_consumer()
	if err != nil {
		fmt.Printf("Error obtaining ProxyPushConsumer: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Got ProxyPushConsumer: %s\n", proxyPushConsumerKey)

	proxyPushConsumer := event.NewCosEvent_ProxyPushConsumer_Stub(client, []byte(proxyPushConsumerKey))

	// 4. Connect this supplier to the ProxyPushConsumer
	err = proxyPushConsumer.Connect_push_supplier("my_supplier")
	if err != nil {
		fmt.Printf("Error connecting push supplier: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Supplier connected successfully")

	// 5. Push events
	messages := []string{
		"Hello from Go supplier!",
		"Event message 2",
		"Event message 3",
	}

	for i, msg := range messages {
		data := cdr.Any{
			Type:  18, // tk_string
			Value: msg,
		}
		err = proxyPushConsumer.Push(data)
		if err != nil {
			fmt.Printf("Error pushing event %d: %v\n", i+1, err)
			os.Exit(1)
		}
		fmt.Printf("Pushed event %d: %s\n", i+1, msg)
		time.Sleep(500 * time.Millisecond)
	}

	// 6. Disconnect
	err = proxyPushConsumer.Disconnect_push_consumer()
	if err != nil {
		fmt.Printf("Error disconnecting: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Supplier disconnected. Done.")
}
