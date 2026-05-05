.PHONY: all build clean idlc generate services run-naming run-event install

IDLC_BIN = bin/idlc
NAMING_BIN = bin/naming-service
EVENT_BIN = bin/event-service

all: build

build: idlc generate services

idlc:
	@echo "Building IDL compiler..."
	@mkdir -p bin
	go build -o $(IDLC_BIN) cmd/idlc/main.go

generate: idlc
	@echo "Generating Go code from IDL..."
	@cd services/naming && ../../$(IDLC_BIN) -pkg naming CosNaming.idl
	@cd services/event && ../../$(IDLC_BIN) -pkg event CosEvent.idl

services:
	@echo "Building CORBA services..."
	@mkdir -p bin
	go build -o $(NAMING_BIN) cmd/naming-service/main.go
	go build -o $(EVENT_BIN) cmd/event-service/main.go

clean:
	@echo "Cleaning binaries and generated code..."
	rm -rf bin/
	rm -f services/naming/*_corba.go
	rm -f services/event/*_corba.go

run-naming:
	@echo "Starting Naming Service..."
	go run cmd/naming-service/main.go

run-event:
	@echo "Starting Event Service..."
	go run cmd/event-service/main.go

install: build
	@echo "Installing binaries to $${GOPATH:-$$HOME/go}/bin..."
	@mkdir -p $${GOPATH:-$$HOME/go}/bin
	cp $(IDLC_BIN) $${GOPATH:-$$HOME/go}/bin/
	cp $(NAMING_BIN) $${GOPATH:-$$HOME/go}/bin/
	cp $(EVENT_BIN) $${GOPATH:-$$HOME/go}/bin/
	@echo "Install complete!"
