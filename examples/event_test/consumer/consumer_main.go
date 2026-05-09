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

// consumer_main.go demonstrates a CosEvent push consumer that connects to
// the EventService, obtains a ProxyPushSupplier, and registers as a consumer.
package main

import (
	"fmt"
	"os"

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

	// 2. Get ConsumerAdmin from EventChannel
	consumerAdminKey, err := eventChannel.For_consumers()
	if err != nil {
		fmt.Printf("Error getting ConsumerAdmin: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Got ConsumerAdmin: %s\n", consumerAdminKey)

	consumerAdmin := event.NewCosEvent_ConsumerAdmin_Stub(client, []byte(consumerAdminKey))

	// 3. Obtain a ProxyPushSupplier from ConsumerAdmin
	proxyPushSupplierKey, err := consumerAdmin.Obtain_push_supplier()
	if err != nil {
		fmt.Printf("Error obtaining ProxyPushSupplier: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Got ProxyPushSupplier: %s\n", proxyPushSupplierKey)

	proxyPushSupplier := event.NewCosEvent_ProxyPushSupplier_Stub(client, []byte(proxyPushSupplierKey))

	// 4. Connect this consumer to the ProxyPushSupplier
	err = proxyPushSupplier.Connect_push_consumer("my_consumer")
	if err != nil {
		fmt.Printf("Error connecting push consumer: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Consumer registered successfully")
	fmt.Println("Consumer is now connected. Events pushed by suppliers will be broadcast to this consumer.")

	// 5. Disconnect when done
	err = proxyPushSupplier.Disconnect_push_supplier()
	if err != nil {
		fmt.Printf("Error disconnecting: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Consumer disconnected. Done.")
}
