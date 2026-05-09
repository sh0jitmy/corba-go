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

	"github.com/sh0jitmy/corba-go/orb/iiop"
	"github.com/sh0jitmy/corba-go/services/naming"
)

func main() {
	server, err := iiop.NewServer(2809)
	if err != nil {
		log.Fatalf("Failed to start naming server: %v", err)
	}

	impl := naming.NewNamingContextImpl()
	handler := naming.CosNaming_NamingContext_Skeleton(impl)

	log.Println("NamingService started on :2809 with ObjectKey 'NameService'")

	err = server.Serve(func(objectKey []byte, operation string, reqPayload []byte) ([]byte, error) {
		if string(objectKey) != "NameService" {
			return nil, nil // Ignored or throw NO_IMPLEMENT
		}
		return handler(objectKey, operation, reqPayload)
	})
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
