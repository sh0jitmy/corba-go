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
