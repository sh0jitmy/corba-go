package main

import (
	"flag"
	"fmt"
	"io/ioutil"
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

	idlFile := flag.Arg(0)
	data, err := ioutil.ReadFile(idlFile)
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
	outFile := name + "_corba.go"

	err = ioutil.WriteFile(outFile, []byte(code), 0644)
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
