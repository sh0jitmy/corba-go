# corba-go
[![Go Reference](https://pkg.go.dev/badge/github.com/sh0jitmy/corba-go.svg)](https://pkg.go.dev/github.com/sh0jitmy/corba-go)
[![Tests](https://github.com/sh0jitmy/corba-go/workflows/Tests/badge.svg)](https://github.com/sh0jitmy/corba-go/actions/workflows/tests.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sh0jitmy/corba-go)](https://goreportcard.com/report/github.com/sh0jitmy/corba-go)


*Read this in other languages: [日本語 (Japanese)](README_ja.md)*

A Native Go implementation of CORBA middleware. This project provides a basic but functional Object Request Broker (ORB) and an IDL compiler (`idlc`) built entirely in Go, aiming to achieve interoperability with conventional CORBA systems.

## Features
- **IDL Compiler (`idlc`)**: Compiles OMG IDL into Go interfaces, client stubs, and server skeletons.
  - Supported IDL types: `struct`, `exception`, `typedef`, `sequence`, `enum`, `union`, basic types (`long`, `short`, `octet`, `string`, `boolean`, etc.), `Object`, and `any`.
  - Generated code is split into three files: `*_types.go` (type definitions), `*_stub.go` (client stubs), `*_skel.go` (server skeletons).
- **ORB Core**: 
  - CDR (Common Data Representation) marshaling/unmarshaling with strict memory alignment handling.
  - GIOP 1.2 & IIOP (Internet Inter-ORB Protocol) support.
  - IOR (Interoperable Object Reference) encoding/decoding utilities.
- **Naming Service (`CosNaming`)**: In-memory name-to-IOR binding and resolution.
- **Event Service (`CosEvent`)**: Push-model event channel for broadcasting messages between suppliers and consumers.
- **REST-CORBA Gateway**: Dynamic JSON-to-CORBA proxy utilizing runtime IDL parsing (DII-like mechanism).

## Directory Structure

```
corba-go/
├── cmd/                          # Command-line entry points
│   ├── idlc/                     # IDL compiler
│   ├── naming-service/           # Standalone CosNaming service
│   ├── event-service/            # Standalone CosEvent service
│   └── rest-gateway/             # Dynamic REST-to-CORBA proxy gateway
├── orb/                          # Core ORB protocols
│   ├── cdr/                      # CDR encoder/decoder + Any type
│   ├── giop/                     # GIOP message handling
│   ├── iiop/                     # IIOP client/server
│   └── iop/                      # IOR encoding/decoding
├── idlc/                         # IDL compiler internals (lexer, parser, AST, generator)
├── services/                     # Standard CORBA service implementations
│   ├── naming/                   # CosNaming (IDL + server + generated code)
│   └── event/                    # CosEvent (IDL + server + generated code)
└── examples/                     # Usage examples
    ├── basic_test/               # Basic IIOP client/server
    ├── naming_test/              # Naming service integration test
    │   └── client/               # Naming client (resolve + invoke)
    └── event_test/               # Event service integration test
        ├── supplier/             # Push supplier
        └── consumer/             # Push consumer
```

## Quick Start (Makefile)

```bash
# Build the IDL compiler (idlc) and all services
make build

# Generate Go code from service IDL files
make generate

# Clean generated binaries and code
make clean

# Run the Naming Service
make run-naming

# Run the Event Service
make run-event
```

## Manual Usage

### 1. Compiling IDL

The `idlc` compiler generates three separate files from each IDL definition:

```bash
# Build the compiler
go build -o bin/idlc ./cmd/idlc/

# Generate Go code from an IDL file
# Output: MyService_types.go, MyService_stub.go, MyService_skel.go
./bin/idlc -pkg mypkg MyService.idl
```

**Options:**
| Flag | Default | Description |
|------|---------|-------------|
| `-pkg` | lowercase IDL module name | Go package name for generated code |
| `-out` | `.` | Output directory for generated files |

### 2. Running Services

```bash
# Start Naming Service (listens on :2809)
go run cmd/naming-service/main.go

# Start Event Service (listens on :2810)
go run cmd/event-service/main.go
```

### 3. IOR Utility

Use the `iop` package for creating and parsing IOR strings:

```go
import "github.com/sh0jitmy/corba-go/orb/iop"

// Create an IOR
ior := iop.NewIOR("IDL:MyModule/MyInterface:1.0", "localhost", 2809, []byte("MyObject"))
iorStr := ior.StringifyIOR()

// Parse an IOR string
parsed, _ := iop.ParseIOR(iorStr)
profile, _ := iop.ParseIIOPProfile(parsed.Profiles[0].ProfileData)
fmt.Printf("Host: %s, Port: %d, ObjectKey: %s\n", profile.Host, profile.Port, string(profile.ObjectKey))
```

### 4. REST-CORBA Gateway

A dynamic gateway that parses an IDL file at runtime and proxies JSON HTTP requests to CORBA IIOP calls without requiring pre-compiled stubs.

```bash
# Build the gateway
go build -o bin/rest-gateway ./cmd/rest-gateway/

# Run the gateway with a target IDL file
./bin/rest-gateway -idl ./examples/basic_test/test.idl -port 8080
```

Send a JSON request to invoke a CORBA method:

```bash
curl -X POST http://localhost:8080/api/invoke \
  -H "Content-Type: application/json" \
  -d '{
    "target_name": [
      {"id": "Math", "kind": "Service"}
    ],
    "interface": "Math",
    "method": "Add",
    "args": [10, 20]
  }'
```

## Examples

### Naming Service Test

Demonstrates name registration and resolution via CosNaming:

```bash
# Terminal 1: Start the Naming Service
./bin/naming-service

# Terminal 2: Start a server that registers itself with the Naming Service
cd examples/naming_test
go run server_main.go test_types.go test_stub.go test_skel.go

# Terminal 3: Resolve the name and invoke operations
cd examples/naming_test/client
go run client_main.go test_types.go test_stub.go test_skel.go
```

### Event Service Test

Demonstrates push-model event broadcasting via CosEvent:

```bash
# Terminal 1: Start the Event Service
./bin/event-service

# Terminal 2: Register a consumer
cd examples/event_test/consumer
go run consumer_main.go

# Terminal 3: Push events from a supplier
cd examples/event_test/supplier
go run supplier_main.go
```

The supplier pushes `cdr.Any` typed events through the EventChannel. The service broadcasts each event to all registered consumers.

## License
This project is published under [Apache 2.0 License](LICENSE).
