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
	"github.com/sh0jitmy/corba-go/services/event"
)

func main() {
	server, err := iiop.NewServer(2810)
	if err != nil {
		log.Fatalf("Failed to start event server: %v", err)
	}

	impl := event.NewEventChannelImpl()

	// Create skeleton handlers for each interface the impl satisfies
	eventChannelHandler := event.CosEvent_EventChannel_Skeleton(impl)
	consumerAdminHandler := event.CosEvent_ConsumerAdmin_Skeleton(impl)
	supplierAdminHandler := event.CosEvent_SupplierAdmin_Skeleton(impl)
	proxyPushConsumerHandler := event.CosEvent_ProxyPushConsumer_Skeleton(impl)
	proxyPushSupplierHandler := event.CosEvent_ProxyPushSupplier_Skeleton(impl)

	log.Println("EventService started on :2810")

	err = server.Serve(func(objectKey []byte, operation string, reqPayload []byte) ([]byte, error) {
		key := string(objectKey)
		log.Printf("EventService: objectKey=%s operation=%s", key, operation)

		switch key {
		case "EventChannel":
			return eventChannelHandler(objectKey, operation, reqPayload)
		case "consumer_admin":
			return consumerAdminHandler(objectKey, operation, reqPayload)
		case "supplier_admin":
			return supplierAdminHandler(objectKey, operation, reqPayload)
		case "proxy_push_consumer":
			return proxyPushConsumerHandler(objectKey, operation, reqPayload)
		case "proxy_push_supplier":
			return proxyPushSupplierHandler(objectKey, operation, reqPayload)
		default:
			return nil, fmt.Errorf("unknown object key: %s", key)
		}
	})
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
