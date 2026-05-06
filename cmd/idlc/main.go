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

	"github.com/shjtmy/corba-go/idlc"
)

func main() {
	pkgName := flag.String("pkg", "main", "Go package name for generated code")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: idlc [-pkg <pkgname>] <file.idl>")
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

	code := idlc.GenerateGoCode(mod, *pkgName)

	base := filepath.Base(idlFile)
	name := stringsTrimSuffix(base, filepath.Ext(base))
	outFile := filepath.Clean(name + "_corba.go")

	err = os.WriteFile(outFile, []byte(code), 0600) //nolint:gosec
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing generated code: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s\n", outFile)
}

func stringsTrimSuffix(s, suffix string) string {
	if len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix {
		return s[:len(s)-len(suffix)]
	}
	return s
}
