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

package idlc

import (
	"fmt"
	"unicode"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenModule
	TokenStruct
	TokenTypedef
	TokenEnum
	TokenUnion
	TokenSwitch
	TokenCase
	TokenDefault
	TokenException
	TokenRaises
	TokenObject
	TokenVoid
	TokenInterface
	TokenIn
	TokenOut
	TokenInout
	TokenSequence
	TokenAny
	TokenOctet
	TokenLBrace // {
	TokenRBrace // }
	TokenLParen // (
	TokenRParen // )
	TokenLAngle // <
	TokenRAngle // >
	TokenSemi   // ;
	TokenComma  // ,
	TokenColon  // :
)

type Token struct {
	Type  TokenType
	Value string
}

type Lexer struct {
	input string
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input, pos: 0}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF, Value: ""}
	}

	ch := l.input[l.pos]

	switch ch {
	case '{':
		l.pos++
		return Token{Type: TokenLBrace, Value: "{"}
	case '}':
		l.pos++
		return Token{Type: TokenRBrace, Value: "}"}
	case '(':
		l.pos++
		return Token{Type: TokenLParen, Value: "("}
	case ')':
		l.pos++
		return Token{Type: TokenRParen, Value: ")"}
	case '<':
		l.pos++
		return Token{Type: TokenLAngle, Value: "<"}
	case '>':
		l.pos++
		return Token{Type: TokenRAngle, Value: ">"}
	case ';':
		l.pos++
		return Token{Type: TokenSemi, Value: ";"}
	case ',':
		l.pos++
		return Token{Type: TokenComma, Value: ","}
	case ':':
		l.pos++
		return Token{Type: TokenColon, Value: ":"}
	}

	if unicode.IsLetter(rune(ch)) || ch == '_' {
		start := l.pos
		for l.pos < len(l.input) && (unicode.IsLetter(rune(l.input[l.pos])) || unicode.IsDigit(rune(l.input[l.pos])) || l.input[l.pos] == '_') {
			l.pos++
		}
		val := l.input[start:l.pos]

		switch val {
		case "module":
			return Token{Type: TokenModule, Value: val}
		case "struct":
			return Token{Type: TokenStruct, Value: val}
		case "typedef":
			return Token{Type: TokenTypedef, Value: val}
		case "enum":
			return Token{Type: TokenEnum, Value: val}
		case "union":
			return Token{Type: TokenUnion, Value: val}
		case "switch":
			return Token{Type: TokenSwitch, Value: val}
		case "case":
			return Token{Type: TokenCase, Value: val}
		case "default":
			return Token{Type: TokenDefault, Value: val}
		case "exception":
			return Token{Type: TokenException, Value: val}
		case "raises":
			return Token{Type: TokenRaises, Value: val}
		case "Object":
			return Token{Type: TokenObject, Value: val}
		case "void":
			return Token{Type: TokenVoid, Value: val}
		case "interface":
			return Token{Type: TokenInterface, Value: val}
		case "in":
			return Token{Type: TokenIn, Value: val}
		case "out":
			return Token{Type: TokenOut, Value: val}
		case "inout":
			return Token{Type: TokenInout, Value: val}
		case "sequence":
			return Token{Type: TokenSequence, Value: val}
		case "any":
			return Token{Type: TokenAny, Value: val}
		case "octet":
			return Token{Type: TokenOctet, Value: val}
		}

		return Token{Type: TokenIdent, Value: val}
	}

	// For simplicity, failing on unexpected characters
	panic(fmt.Sprintf("Unexpected character: %c", ch))
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsSpace(rune(ch)) {
			l.pos++
		} else {
			break
		}
	}
}
