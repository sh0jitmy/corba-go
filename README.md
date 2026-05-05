# corba-go

*Read this in other languages: [日本語 (Japanese)](README_ja.md)*

A Native Go implementation of CORBA middleware. This project provides a basic but functional Object Request Broker (ORB) and an IDL compiler (`idlc`) built entirely in Go, enabling interoperability with legacy CORBA systems (e.g., omniORB, JacORB).

## Features
- **IDL Compiler (`idlc`)**: Compiles OMG IDL into Go interfaces, client stubs, and server skeletons.
  - Supported IDL types: `struct`, `exception`, `typedef`, `sequence`, `enum`, `union`, basic types (`long`, `short`, `octet`, `string`, `boolean`, etc.), and `Object`.
- **ORB Core**: 
  - CDR (Common Data Representation) marshaling/unmarshaling with strict memory alignment handling.
  - GIOP 1.2 & IIOP (Internet Inter-ORB Protocol) support.
- **Naming Service (`CosNaming`)**: In-memory name-to-IOR binding and resolution.
- **Event Service (`CosEvent`)**: Simple push-model event channel for broadcasting messages.

## Directory Structure
- `cmd/`: Command-line entry points.
  - `idlc/`: The IDL compiler.
  - `naming-service/`: The standalone CosNaming service.
  - `event-service/`: The standalone CosEvent service.
- `orb/`: Core ORB protocols (`cdr`, `giop`, `iiop`, `iop`).
- `services/`: Implementations of standard CORBA services (`naming`, `event`).
- `idlc/`: Parser, AST, and Go code generator packages for the IDL compiler.

## Quick Start (Makefile)

We provide a `Makefile` to simplify building and running the project.

```bash
# Build the IDL compiler (idlc) and all services
make build

# Clean generated binaries and stubs
make clean

# Run the Naming Service
make run-naming

# Run the Event Service
make run-event
```

## Manual Usage

### 1. Compiling IDL
```bash
# Build the compiler
go build -o bin/idlc cmd/idlc/main.go

# Generate Go code from an IDL file
./bin/idlc -pkg mypkg my_interface.idl
```

### 2. Running Services
```bash
# Start Naming Service (listens on :2809)
go run cmd/naming-service/main.go

# Start Event Service (listens on :2810)
go run cmd/event-service/main.go
```
