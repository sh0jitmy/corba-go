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

package event

import (
	"fmt"
	"sync"

	"github.com/sh0jitmy/corba-go/orb/cdr"
)

type EventChannelImpl struct {
	consumers []string
	mu        sync.RWMutex
}

func NewEventChannelImpl() *EventChannelImpl {
	return &EventChannelImpl{
		consumers: make([]string, 0),
	}
}

// EventChannel methods
func (e *EventChannelImpl) For_consumers() (string, error) {
	// returns ConsumerAdmin object key (for simplicity we use a dummy logic)
	return "consumer_admin", nil
}

func (e *EventChannelImpl) For_suppliers() (string, error) {
	// returns SupplierAdmin object key
	return "supplier_admin", nil
}

func (e *EventChannelImpl) Destroy() error {
	return nil
}

// For simplicity in this dummy setup, we implement the ProxyPushSupplier directly here
// as well as ProxyPushConsumer. In reality they would be distinct objects.
func (e *EventChannelImpl) Connect_push_consumer(push_consumer string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.consumers = append(e.consumers, push_consumer)
	fmt.Printf("EventChannel: Connected push consumer %s\n", push_consumer)
	return nil
}

func (e *EventChannelImpl) Disconnect_push_supplier() error {
	return nil
}

func (e *EventChannelImpl) Connect_push_supplier(push_supplier string) error {
	fmt.Printf("EventChannel: Connected push supplier %s\n", push_supplier)
	return nil
}

func (e *EventChannelImpl) Push(data cdr.Any) error {
	e.mu.RLock()
	subs := make([]string, len(e.consumers))
	copy(subs, e.consumers)
	e.mu.RUnlock()

	fmt.Printf("EventChannel: Received push data (Type: %v), broadcasting to %d consumers...\n", data.Type, len(subs))
	for _, sub := range subs {
		fmt.Printf(" -> Pushed to consumer %s\n", sub)
	}
	return nil
}

func (e *EventChannelImpl) Disconnect_push_consumer() error {
	return nil
}

// SupplierAdmin methods
func (e *EventChannelImpl) Obtain_push_consumer() (string, error) {
	return "proxy_push_consumer", nil
}

// ConsumerAdmin methods
func (e *EventChannelImpl) Obtain_push_supplier() (string, error) {
	return "proxy_push_supplier", nil
}
