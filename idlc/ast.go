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

// AST Nodes for simplified IDL

type ASTNode interface{}

type ModuleNode struct {
	Name       string
	Typedefs   []*TypedefNode
	Enums      []*EnumNode
	Unions     []*UnionNode
	Exceptions []*ExceptionNode
	Structs    []*StructNode
	Interfaces []*InterfaceNode
}

type TypedefNode struct {
	Type string
	Name string
}

type EnumNode struct {
	Name   string
	Values []string
}

type UnionNode struct {
	Name       string
	SwitchType string
	Cases      []*UnionCaseNode
}

type UnionCaseNode struct {
	IsDefault bool
	Labels    []string // e.g. "1", "2", "TRUE"
	Type      string
	Name      string
}

type StructNode struct {
	Name   string
	Fields []*FieldNode
}

type ExceptionNode struct {
	Name   string
	Fields []*FieldNode
}

type FieldNode struct {
	Type string
	Name string
}

type InterfaceNode struct {
	Name       string
	Operations []*OperationNode
}

type OperationNode struct {
	Name       string
	ReturnType string
	Parameters []*ParameterNode
}

type ParameterNode struct {
	Direction string // "in", "out", "inout"
	Type      string
	Name      string
}
