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

import "fmt"

type Parser struct {
	lexer *Lexer
	cur   Token
	next  Token
}

func NewParser(lexer *Lexer) *Parser {
	p := &Parser{lexer: lexer}
	p.advance()
	p.advance()
	return p
}

func (p *Parser) advance() {
	p.cur = p.next
	p.next = p.lexer.NextToken()
}

func (p *Parser) expect(t TokenType) error {
	if p.cur.Type == t {
		p.advance()
		return nil
	}
	return fmt.Errorf("expected token %v, got %v", t, p.cur.Type)
}

func (p *Parser) Parse() (*ModuleNode, error) {
	return p.parseModule()
}

func (p *Parser) parseModule() (*ModuleNode, error) {
	if err := p.expect(TokenModule); err != nil {
		return nil, err
	}
	nameTok := p.cur
	if err := p.expect(TokenIdent); err != nil {
		return nil, err
	}

	mod := &ModuleNode{Name: nameTok.Value}

	if err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	for p.cur.Type != TokenRBrace && p.cur.Type != TokenEOF {
		switch p.cur.Type {
		case TokenInterface:
			iface, err := p.parseInterface()
			if err != nil {
				return nil, err
			}
			mod.Interfaces = append(mod.Interfaces, iface)
		case TokenStruct:
			st, err := p.parseStruct()
			if err != nil {
				return nil, err
			}
			mod.Structs = append(mod.Structs, st)
		case TokenTypedef:
			td, err := p.parseTypedef()
			if err != nil {
				return nil, err
			}
			mod.Typedefs = append(mod.Typedefs, td)
		case TokenEnum:
			en, err := p.parseEnum()
			if err != nil {
				return nil, err
			}
			mod.Enums = append(mod.Enums, en)
		case TokenUnion:
			un, err := p.parseUnion()
			if err != nil {
				return nil, err
			}
			mod.Unions = append(mod.Unions, un)
		case TokenException:
			ex, err := p.parseException()
			if err != nil {
				return nil, err
			}
			mod.Exceptions = append(mod.Exceptions, ex)
		default:
			return nil, fmt.Errorf("unexpected token in module: %v", p.cur.Value)
		}
	}

	if err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	if p.cur.Type == TokenSemi {
		p.advance()
	}

	return mod, nil
}

func (p *Parser) parseStruct() (*StructNode, error) {
	if err := p.expect(TokenStruct); err != nil {
		return nil, err
	}
	nameTok := p.cur
	if err := p.expect(TokenIdent); err != nil {
		return nil, err
	}

	st := &StructNode{Name: nameTok.Value}

	if err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	for p.cur.Type != TokenRBrace && p.cur.Type != TokenEOF {
		fieldType, err := p.parseType()
		if err != nil {
			return nil, err
		}
		nameTok := p.cur
		if err := p.expect(TokenIdent); err != nil {
			return nil, err
		}

		if err := p.expect(TokenSemi); err != nil {
			return nil, err
		}

		st.Fields = append(st.Fields, &FieldNode{
			Type: fieldType,
			Name: nameTok.Value,
		})
	}

	if err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}
	if err := p.expect(TokenSemi); err != nil {
		return nil, err
	}

	return st, nil
}

func (p *Parser) parseInterface() (*InterfaceNode, error) {
	if err := p.expect(TokenInterface); err != nil {
		return nil, err
	}

	ifaceName := p.cur.Value
	if err := p.expect(TokenIdent); err != nil {
		return nil, err
	}

	if err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	iface := &InterfaceNode{Name: ifaceName}

	for p.cur.Type != TokenRBrace && p.cur.Type != TokenEOF {
		op, err := p.parseOperation()
		if err != nil {
			return nil, err
		}
		iface.Operations = append(iface.Operations, op)
	}

	if err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}
	if err := p.expect(TokenSemi); err != nil {
		return nil, err
	}

	return iface, nil
}

func (p *Parser) parseOperation() (*OperationNode, error) {
	returnType, err := p.parseType()
	if err != nil {
		return nil, err
	}

	opName := p.cur.Value
	if err := p.expect(TokenIdent); err != nil {
		return nil, err
	}

	if err := p.expect(TokenLParen); err != nil {
		return nil, err
	}

	op := &OperationNode{
		Name:       opName,
		ReturnType: returnType,
	}

	for p.cur.Type != TokenRParen && p.cur.Type != TokenEOF {
		dir := p.cur.Value
		switch p.cur.Type {
		case TokenIn:
			if err := p.expect(TokenIn); err != nil {
				return nil, err
			}
		case TokenOut:
			if err := p.expect(TokenOut); err != nil {
				return nil, err
			}
		case TokenInout:
			if err := p.expect(TokenInout); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("expected direction (in/out/inout), got %v", p.cur.Value)
		}

		paramType, err := p.parseType()
		if err != nil {
			return nil, err
		}

		paramName := p.cur.Value
		if err := p.expect(TokenIdent); err != nil {
			return nil, err
		}

		op.Parameters = append(op.Parameters, &ParameterNode{
			Direction: dir,
			Type:      paramType,
			Name:      paramName,
		})

		if p.cur.Type == TokenComma {
			p.advance()
		}
	}

	if err := p.expect(TokenRParen); err != nil {
		return nil, err
	}

	// handle optional raises(...)
	if p.cur.Type == TokenRaises {
		p.advance()
		if err := p.expect(TokenLParen); err != nil {
			return nil, err
		}
		for p.cur.Type != TokenRParen && p.cur.Type != TokenEOF {
			// skip exception names
			p.advance()
			if p.cur.Type == TokenComma {
				p.advance()
			}
		}
		if err := p.expect(TokenRParen); err != nil {
			return nil, err
		}
	}

	if err := p.expect(TokenSemi); err != nil {
		return nil, err
	}

	return op, nil
}

func (p *Parser) parseType() (string, error) {
	switch p.cur.Type {
	case TokenIdent, TokenAny, TokenOctet, TokenObject, TokenVoid:
		t := p.cur.Value
		p.advance()
		return t, nil
	case TokenSequence:
		p.advance()
		if err := p.expect(TokenLAngle); err != nil {
			return "", err
		}
		inner, err := p.parseType()
		if err != nil {
			return "", err
		}
		if err := p.expect(TokenRAngle); err != nil {
			return "", err
		}
		return "sequence<" + inner + ">", nil
	}
	return "", fmt.Errorf("expected type, got %v", p.cur.Value)
}

func (p *Parser) parseTypedef() (*TypedefNode, error) {
	if err := p.expect(TokenTypedef); err != nil {
		return nil, err
	}
	typ, err := p.parseType()
	if err != nil {
		return nil, err
	}
	nameTok := p.cur
	if err := p.expect(TokenIdent); err != nil {
		return nil, err
	}
	if err := p.expect(TokenSemi); err != nil {
		return nil, err
	}
	return &TypedefNode{Type: typ, Name: nameTok.Value}, nil
}

func (p *Parser) parseEnum() (*EnumNode, error) {
	if err := p.expect(TokenEnum); err != nil {
		return nil, err
	}
	nameTok := p.cur
	if err := p.expect(TokenIdent); err != nil {
		return nil, err
	}
	if err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	en := &EnumNode{Name: nameTok.Value}
	for p.cur.Type != TokenRBrace && p.cur.Type != TokenEOF {
		valTok := p.cur
		if err := p.expect(TokenIdent); err != nil {
			return nil, err
		}
		en.Values = append(en.Values, valTok.Value)

		if p.cur.Type == TokenComma {
			p.advance()
		}
	}
	if err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}
	if err := p.expect(TokenSemi); err != nil {
		return nil, err
	}
	return en, nil
}

func (p *Parser) parseUnion() (*UnionNode, error) {
	if err := p.expect(TokenUnion); err != nil {
		return nil, err
	}
	nameTok := p.cur
	if err := p.expect(TokenIdent); err != nil {
		return nil, err
	}
	if err := p.expect(TokenSwitch); err != nil {
		return nil, err
	}
	if err := p.expect(TokenLParen); err != nil {
		return nil, err
	}
	switchType, err := p.parseType()
	if err != nil {
		return nil, err
	}
	if err := p.expect(TokenRParen); err != nil {
		return nil, err
	}
	if err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	un := &UnionNode{Name: nameTok.Value, SwitchType: switchType}

	for p.cur.Type != TokenRBrace && p.cur.Type != TokenEOF {
		cNode := &UnionCaseNode{}

		for p.cur.Type == TokenCase || p.cur.Type == TokenDefault {
			switch p.cur.Type {
			case TokenCase:
				p.advance()
				// Simplified label parsing: treat identifier/number as label string
				labelTok := p.cur
				p.advance() // advance label
				cNode.Labels = append(cNode.Labels, labelTok.Value)
				if err := p.expect(TokenColon); err != nil {
					return nil, err
				}
			case TokenDefault:
				p.advance()
				cNode.IsDefault = true
				if err := p.expect(TokenColon); err != nil {
					return nil, err
				}
			}
		}

		typ, err := p.parseType()
		if err != nil {
			return nil, err
		}
		cNode.Type = typ

		fieldTok := p.cur
		if err := p.expect(TokenIdent); err != nil {
			return nil, err
		}
		cNode.Name = fieldTok.Value

		if err := p.expect(TokenSemi); err != nil {
			return nil, err
		}
		un.Cases = append(un.Cases, cNode)
	}

	if err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}
	if err := p.expect(TokenSemi); err != nil {
		return nil, err
	}
	return un, nil
}

func (p *Parser) parseException() (*ExceptionNode, error) {
	if err := p.expect(TokenException); err != nil {
		return nil, err
	}
	nameTok := p.cur
	if err := p.expect(TokenIdent); err != nil {
		return nil, err
	}

	ex := &ExceptionNode{Name: nameTok.Value}

	if err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	for p.cur.Type != TokenRBrace && p.cur.Type != TokenEOF {
		fieldType, err := p.parseType()
		if err != nil {
			return nil, err
		}
		nameTok := p.cur
		if err := p.expect(TokenIdent); err != nil {
			return nil, err
		}

		if err := p.expect(TokenSemi); err != nil {
			return nil, err
		}

		ex.Fields = append(ex.Fields, &FieldNode{
			Type: fieldType,
			Name: nameTok.Value,
		})
	}

	if err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}
	if err := p.expect(TokenSemi); err != nil {
		return nil, err
	}

	return ex, nil
}
