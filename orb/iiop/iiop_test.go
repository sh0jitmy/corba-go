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

package iiop

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/shjtmy/corba-go/orb/cdr"
)

func TestClientServer(t *testing.T) {
	// Start server on a random port
	server, err := NewServer(0)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	addr := server.Addr().String()
	port := uint16(0)
	// Parse port
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("Failed to split addr: %v", err)
	}
	portVal, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		t.Fatalf("Failed to parse port: %v", err)
	}
	port = uint16(portVal)

	go func() {
		err := server.Serve(func(objectKey []byte, operation string, reqPayload []byte) ([]byte, error) {
			if string(objectKey) != "TestObject" {
				return nil, fmt.Errorf("Unknown object: %s", string(objectKey))
			}

			if operation == "echo" {
				// Read string from payload
				dec := cdr.NewDecoder(reqPayload, binary.LittleEndian)
				// GIOP 1.2 might have alignment paddings at the start of payload, we'll try to decode directly.
				// Since we sent exactly after 8-byte alignment, the string should be the first thing.
				str, err := dec.DecodeString()
				if err != nil {
					return nil, err
				}

				// Encode reply
				enc := cdr.NewEncoder(binary.LittleEndian)
				enc.EncodeString("Echo: " + str)
				return enc.Bytes(), nil
			}

			return nil, fmt.Errorf("Unknown operation: %s", operation)
		})
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			t.Errorf("Server error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Create client
	client, err := NewClient("127.0.0.1", port)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	// Encode request payload
	enc := cdr.NewEncoder(binary.LittleEndian)
	enc.EncodeString("Hello CORBA")

	// Invoke
	reply, err := client.Invoke([]byte("TestObject"), "echo", enc.Bytes())
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	// Decode reply
	dec := cdr.NewDecoder(reply, binary.LittleEndian)
	replyStr, err := dec.DecodeString()
	if err != nil {
		t.Fatalf("Failed to decode reply string: %v", err)
	}

	expected := "Echo: Hello CORBA"
	if replyStr != expected {
		t.Errorf("Expected '%s', got '%s'", expected, replyStr)
	}
}
