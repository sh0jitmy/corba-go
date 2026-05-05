package main

import (
	"log"

	"github.com/shjtmy/corba-go/orb/iiop"
	"github.com/shjtmy/corba-go/services/naming"
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
