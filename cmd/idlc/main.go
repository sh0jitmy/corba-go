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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sh0jitmy/corba-go/idlc"
)

func main() {
	pkgName := flag.String("pkg", "", "Go package name for generated code (default: lowercase IDL module name)")
	outDir := flag.String("out", ".", "Output directory for generated files")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: idlc [-pkg <pkgname>] [-out <dir>] <file.idl>")
		os.Exit(1)
	}

	idlFile := filepath.Clean(flag.Arg(0))
	data, err := os.ReadFile(idlFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	lexer := idlc.NewLexer(string(data))
	parser := idlc.NewParser(lexer)

	mod, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing IDL: %v\n", err)
		os.Exit(1)
	}

	// Derive package name from module name if not specified
	pkg := *pkgName
	if pkg == "" {
		pkg = strings.ToLower(mod.Name)
	}

	base := filepath.Base(idlFile)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	cleanOut := filepath.Clean(*outDir)

	// Generate and write each file
	files := []struct {
		suffix  string
		content string
	}{
		{"_types.go", idlc.GenerateTypes(mod, pkg)},
		{"_stub.go", idlc.GenerateStub(mod, pkg)},
		{"_skel.go", idlc.GenerateSkel(mod, pkg)},
	}

	for _, f := range files {
		outFile := filepath.Clean(filepath.Join(cleanOut, name+f.suffix))
		if err := os.WriteFile(outFile, []byte(f.content), 0600); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outFile, err)
			os.Exit(1)
		}
		fmt.Printf("Generated %s\n", outFile)
	}
}
