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
	"log"

	"github.com/shjtmy/corba-go/orb/iiop"
	"github.com/shjtmy/corba-go/services/event"
)

func main() {
	server, err := iiop.NewServer(2810) // Event channel listens on 2810
	if err != nil {
		log.Fatalf("Failed to start event server: %v", err)
	}

	impl := event.NewEventChannelImpl()
	handler := event.CosEvent_EventChannel_Skeleton(impl)

	log.Println("EventService started on :2810 with ObjectKey 'EventChannel'")

	err = server.Serve(func(objectKey []byte, operation string, reqPayload []byte) ([]byte, error) {
		if string(objectKey) != "EventChannel" {
			return nil, nil // Ignored
		}
		return handler(objectKey, operation, reqPayload)
	})
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
