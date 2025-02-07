package codegen

// Program represents the root of the AST
type Program struct {
	Classes []*Class
}

// Class represents a COOL class
type Class struct {
	Name     string
	Parent   string
	Features []Feature
}

// Feature represents either a method or attribute
type Feature interface {
	isFeature()
}

// Method represents a class method
type Method struct {
	Name       string
	Formals    []*Formal
	ReturnType string
	Body       Expression
}

func (*Method) isFeature() {}

// Attribute represents a class attribute
type Attribute struct {
	Name string
	Type string
	Init Expression
}

func (*Attribute) isFeature() {}

// Formal represents a method parameter
type Formal struct {
	Name string
	Type string
}

// Expression interface for all COOL expressions
type Expression interface {
	isExpression()
}

// IntConstant represents integer literals
type IntConstant struct {
	Value int
}

func (*IntConstant) isExpression() {}

// StringConstant represents string literals
type StringConstant struct {
	Value string
}

func (*StringConstant) isExpression() {}

// BoolConstant represents boolean literals
type BoolConstant struct {
	Value bool
}

func (*BoolConstant) isExpression() {}

// Dispatch represents method calls
type Dispatch struct {
	Object     Expression
	MethodName string
	Arguments  []Expression
}

func (*Dispatch) isExpression() {}

// If represents if-then-else expressions
type If struct {
	Condition  Expression
	ThenBranch Expression
	ElseBranch Expression
}

func (*If) isExpression() {}

// While represents while loops
type While struct {
	Condition Expression
	Body      Expression
}

func (*While) isExpression() {}

// Block represents a sequence of expressions
type Block struct {
	Expressions []Expression
}

func (*Block) isExpression() {}
