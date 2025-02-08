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

// ========== Identifiers ==========
type TypeIdentifier struct {
	Token lexer.Token
	Value string
}

func (ti *TypeIdentifier) TokenLiteral() string { return ti.Token.Literal }
func (ti *TypeIdentifier) expressionNode()      {}

type ObjectIdentifier struct {
	Token lexer.Token
	Value string
}

func (oi *ObjectIdentifier) TokenLiteral() string { return oi.Token.Literal }
func (oi *ObjectIdentifier) expressionNode()      {}

// ========== Program Structure ==========
type Program struct {
	Classes []*Class
}

func (p *Program) TokenLiteral() string {
	if len(p.Classes) > 0 {
		return p.Classes[0].TokenLiteral()
	}
	return ""
}

type Class struct {
	Token    lexer.Token
	Name     *TypeIdentifier
	Parent   *TypeIdentifier // Optional inherits
	Features []Feature
}

func (c *Class) TokenLiteral() string { return c.Token.Literal }
func (c *Class) statementNode()       {}

// ========== Features (Attributes/Methods) ==========
type Attribute struct {
	Token lexer.Token
	Name  *ObjectIdentifier
	Type  *TypeIdentifier
	Init  Expression // Optional initialization
}

func (a *Attribute) TokenLiteral() string { return a.Token.Literal }
func (a *Attribute) featureNode()         {}

type Method struct {
	Token      lexer.Token
	Name       *ObjectIdentifier
	Formals    []*Formal
	ReturnType *TypeIdentifier
	Body       Expression
}

func (m *Method) TokenLiteral() string { return m.Token.Literal }
func (m *Method) featureNode()         {}

type Formal struct {
	Token lexer.Token
	Name  *ObjectIdentifier
	Type  *TypeIdentifier
}

// ========== Expressions ==========
type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) expressionNode()      {}

type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) expressionNode()      {}

type BooleanLiteral struct {
	Token lexer.Token
	Value bool
}

func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) expressionNode()      {}

type Assignment struct {
	Token lexer.Token
	Left  *ObjectIdentifier
	Value Expression
}

func (a *Assignment) TokenLiteral() string { return a.Token.Literal }
func (a *Assignment) expressionNode()      {}

type Dispatch struct {
	Token      lexer.Token
	Receiver   Expression      // Can be nil (implicit self)
	StaticType *TypeIdentifier // Optional for static dispatch (@)
	Method     *ObjectIdentifier
	Args       []Expression
}

func (d *Dispatch) TokenLiteral() string { return d.Token.Literal }
func (d *Dispatch) expressionNode()      {}

type Conditional struct {
	Token      lexer.Token
	Predicate  Expression
	ThenBranch Expression
	ElseBranch Expression
}

func (c *Conditional) TokenLiteral() string { return c.Token.Literal }
func (c *Conditional) expressionNode()      {}

type Loop struct {
	Token     lexer.Token
	Condition Expression
	Body      Expression
}

func (l *Loop) TokenLiteral() string { return l.Token.Literal }
func (l *Loop) expressionNode()      {}

type Block struct {
	Token       lexer.Token
	Expressions []Expression
}

func (b *Block) TokenLiteral() string { return b.Token.Literal }
func (b *Block) expressionNode()      {}

type LetDeclaration struct {
	Token lexer.Token
	Name  *ObjectIdentifier
	Type  *TypeIdentifier
	Init  Expression // Optional
}

type Let struct {
	Token   lexer.Token
	VarName *ObjectIdentifier
	VarType *TypeIdentifier
	VarInit Expression
	Body    Expression
}

func (l *Let) TokenLiteral() string { return l.Token.Literal }
func (l *Let) expressionNode()      {}

type CaseBranch struct {
	Token   lexer.Token
	VarName *ObjectIdentifier
	VarType *TypeIdentifier
	Expr    Expression
}

type Case struct {
	Token    lexer.Token
	Expr     Expression
	Branches []*CaseBranch
}

func (c *Case) TokenLiteral() string { return c.Token.Literal }
func (c *Case) expressionNode()      {}

type New struct {
	Token lexer.Token
	Type  *TypeIdentifier
}

func (n *New) TokenLiteral() string { return n.Token.Literal }
func (n *New) expressionNode()      {}

type IsVoid struct {
	Token lexer.Token
	Expr  Expression
}

func (iv *IsVoid) TokenLiteral() string { return iv.Token.Literal }
func (iv *IsVoid) expressionNode()      {}

type BinaryExpression struct {
	Token    lexer.Token
	Left     Expression
	Operator lexer.TokenType
	Right    Expression
}

func (be *BinaryExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BinaryExpression) expressionNode()      {}

type UnaryExpression struct {
	Token    lexer.Token
	Operator lexer.TokenType
	Right    Expression
}

func (ue *UnaryExpression) TokenLiteral() string { return ue.Token.Literal }
func (ue *UnaryExpression) expressionNode()      {}

type SelfExpression struct {
	Token lexer.Token
}

func (s *SelfExpression) TokenLiteral() string { return s.Token.Literal }
func (s *SelfExpression) expressionNode()      {}
