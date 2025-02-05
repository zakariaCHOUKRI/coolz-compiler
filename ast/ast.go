package ast

import (
	"coolz-compiler/lexer"
)

type Node interface {
	TokenLiteral() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Feature interface {
	Node
	featureNode()
}

type TypeIdentifier struct {
	Token lexer.Token
	Value string
}

func (ti *TypeIdentifier) TokenLiteral() string { return ti.Token.Literal }

type ObjectIdentifier struct {
	Token lexer.Token
	Value string
}

func (oi *ObjectIdentifier) TokenLiteral() string { return oi.Token.Literal }
func (oi *ObjectIdentifier) expressionNode()      {}

type Program struct {
	Classes []*Class
}

func (p *Program) TokenLiteral() string { return "" }

type Class struct {
	Token    lexer.Token
	Name     *TypeIdentifier
	Parent   *TypeIdentifier
	Features []Feature
}

func (c *Class) TokenLiteral() string { return c.Token.Literal }

type Attribute struct {
	Name *ObjectIdentifier
	Type *TypeIdentifier
}

func (a *Attribute) TokenLiteral() string { return a.Name.Value }
func (a *Attribute) featureNode()         {}

type Method struct {
	Name    *ObjectIdentifier // machi logique OnjectID li kan cs == Token itself ??
	Type    *TypeIdentifier
	Formals []*Formal
}

func (m *Method) TokenLiteral() string { return m.Name.Value }
func (m *Method) featureNode()         {}

type Formal struct {
	Name *ObjectIdentifier
	Type *TypeIdentifier
}

func (f *Formal) TokenLiteral() string { return f.Name.Value }

// IntegerLiteral represents an integer literal in the AST.
type IntegerLiteral struct {
	Token lexer.Token // The token representing the integer literal.
	Value int64       // The actual value of the integer literal.
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

// StringLiteral represents a string literal in the AST.
type StringLiteral struct {
	Token lexer.Token // The token representing the string literal.
	Value string      // The actual value of the string literal.
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }

// BooleanLiteral represents a boolean literal in the AST.
type BooleanLiteral struct {
	Token lexer.Token // The token representing the boolean literal.
	Value bool        // The actual value of the boolean literal.
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }

// UnaryExpression represents a unary operation in the AST.
type UnaryExpression struct {
	Token    lexer.Token // The operator token, e.g., 'not', '~', 'isvoid'.
	Operator string      // The operator as a string.
	Right    Expression  // The right-hand side expression.
}

func (ue *UnaryExpression) expressionNode()      {}
func (ue *UnaryExpression) TokenLiteral() string { return ue.Token.Literal }

// BinaryExpression represents a binary operation in the AST.
type BinaryExpression struct {
	Token    lexer.Token // The operator token, e.g., '+', '-', '*', '/'.
	Operator string      // The operator as a string.
	Left     Expression  // The left-hand side expression.
	Right    Expression  // The right-hand side expression.
}

func (be *BinaryExpression) expressionNode()      {}
func (be *BinaryExpression) TokenLiteral() string { return be.Token.Literal }

// IfExpression represents an if-else expression in the AST.
type IfExpression struct {
	Token       lexer.Token // The 'if' token.
	Condition   Expression  // The condition expression.
	Consequence Expression  // The consequence expression (then branch).
	Alternative Expression  // The alternative expression (else branch).
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }

// WhileExpression represents a while loop in the AST.
type WhileExpression struct {
	Token     lexer.Token // The 'while' token.
	Condition Expression  // The condition expression.
	Body      Expression  // The body expression.
}

func (we *WhileExpression) expressionNode()      {}
func (we *WhileExpression) TokenLiteral() string { return we.Token.Literal }

// BlockExpression represents a block of expressions in the AST.
type BlockExpression struct {
	Token       lexer.Token  // The '{' token.
	Expressions []Expression // The list of expressions within the block.
}

func (be *BlockExpression) expressionNode()      {}
func (be *BlockExpression) TokenLiteral() string { return be.Token.Literal }

// LetExpression represents a let expression in the AST.
type LetExpression struct {
	Token    lexer.Token   // The 'let' token.
	Bindings []*LetBinding // The list of bindings (variable declarations).
	In       Expression    // The expression that follows the bindings.
}

func (le *LetExpression) expressionNode()      {}
func (le *LetExpression) TokenLiteral() string { return le.Token.Literal }

// LetBinding represents a single binding in a let expression.
type LetBinding struct {
	Identifier *ObjectIdentifier // The identifier of the binding.
	Type       *TypeIdentifier   // The type of the binding.
	Init       Expression        // The initialization expression, if any.
}

// NewExpression represents the 'new' type expression in the AST.
type NewExpression struct {
	Token lexer.Token     // The 'new' token.
	Type  *TypeIdentifier // The type to be instantiated.
}

func (ne *NewExpression) expressionNode()      {}
func (ne *NewExpression) TokenLiteral() string { return ne.Token.Literal }

// IsVoidExpression represents an 'isvoid' expression in the AST.
type IsVoidExpression struct {
	Token      lexer.Token // The 'isvoid' token.
	Expression Expression  // The expression to check for being void.
}

func (ive *IsVoidExpression) expressionNode()      {}
func (ive *IsVoidExpression) TokenLiteral() string { return ive.Token.Literal }

// AssignExpression represents an assignment expression in the AST.
type AssignExpression struct {
	Token    lexer.Token // The token.ASSIGN token
	Left     Expression
	Operator string
	Right    Expression
}

func (ae *AssignExpression) expressionNode()      {}
func (ae *AssignExpression) TokenLiteral() string { return ae.Token.Literal }
