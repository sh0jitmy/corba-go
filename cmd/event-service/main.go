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
