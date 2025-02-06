package builder

import (
	"coolz-compiler/ast"
	"coolz-compiler/lexer"
	"fmt"
)

// Builder provides a fluent API for constructing ASTs
type Builder struct {
	program *ast.Program
}

// NewBuilder creates a new AST builder
func NewBuilder() *Builder {
	return &Builder{
		program: &ast.Program{},
	}
}

// Class starts building a new class
func (b *Builder) Class(name string) *ClassBuilder {
	class := &ast.Class{
		Name: &ast.TypeIdentifier{
			Token: lexer.Token{Type: lexer.TYPEID, Literal: name},
			Value: name,
		},
	}
	b.program.Classes = append(b.program.Classes, class)
	return &ClassBuilder{class: class}
}

// Build returns the constructed Program
func (b *Builder) Build() *ast.Program {
	return b.program
}

// ClassBuilder builds a class definition
type ClassBuilder struct {
	class *ast.Class
}

// Inherits specifies the parent class
func (cb *ClassBuilder) Inherits(parent string) *ClassBuilder {
	cb.class.Parent = &ast.TypeIdentifier{
		Token: lexer.Token{Type: lexer.TYPEID, Literal: parent},
		Value: parent,
	}
	return cb
}

// Method adds a method to the class
func (cb *ClassBuilder) Method(name string) *MethodBuilder {
	method := &ast.Method{
		Name: &ast.ObjectIdentifier{
			Token: lexer.Token{Type: lexer.OBJECTID, Literal: name},
			Value: name,
		},
	}
	cb.class.Features = append(cb.class.Features, method)
	return &MethodBuilder{method: method}
}

// Attribute adds an attribute to the class
func (cb *ClassBuilder) Attribute(name string) *AttributeBuilder {
	attr := &ast.Attribute{
		Name: &ast.ObjectIdentifier{
			Token: lexer.Token{Type: lexer.OBJECTID, Literal: name},
			Value: name,
		},
	}
	cb.class.Features = append(cb.class.Features, attr)
	return &AttributeBuilder{attr: attr}
}

// AttributeBuilder builds class attributes
type AttributeBuilder struct {
	attr *ast.Attribute
}

// Type sets the attribute type
func (ab *AttributeBuilder) Type(typeName string) *AttributeBuilder {
	ab.attr.Type = &ast.TypeIdentifier{
		Token: lexer.Token{Type: lexer.TYPEID, Literal: typeName},
		Value: typeName,
	}
	return ab
}

// Init sets the initialization expression for the attribute
func (ab *AttributeBuilder) Init(init ast.Expression) *AttributeBuilder {
	ab.attr.Init = init
	return ab
}

// Build returns the completed attribute
func (ab *AttributeBuilder) Build() *ast.Attribute {
	return ab.attr
}

// Continue implementing builders for other AST components...

// Int creates an integer literal (moved directly to Builder)
func (b *Builder) Int(value int64) ast.Expression {
	return &ast.IntegerLiteral{
		Token: lexer.Token{Type: lexer.INT_CONST, Literal: fmt.Sprintf("%d", value)},
		Value: value,
	}
}

// String creates a string literal (moved directly to Builder)
func (b *Builder) String(value string) ast.Expression {
	return &ast.StringLiteral{
		Token: lexer.Token{Type: lexer.STR_CONST, Literal: value},
		Value: value,
	}
}

// Bool creates a boolean literal
func (b *Builder) Bool(value bool) ast.Expression {
	return &ast.BooleanLiteral{
		Token: lexer.Token{Type: lexer.BOOL_CONST, Literal: fmt.Sprintf("%t", value)},
		Value: value,
	}
}

// Binary creates a binary expression
func (b *Builder) Binary(left ast.Expression, operator string, right ast.Expression) ast.Expression {
	return &ast.BinaryExpression{
		Token:    lexer.Token{Type: lexer.GetOperatorType(operator), Literal: operator},
		Left:     left,
		Right:    right,
		Operator: operator,
	}
}

// If creates an if expression
func (b *Builder) If(condition ast.Expression) *IfBuilder {
	return &IfBuilder{
		ifExp: &ast.IfExpression{
			Token:     lexer.Token{Type: lexer.IF},
			Condition: condition,
		},
	}
}

type IfBuilder struct {
	ifExp *ast.IfExpression
}

func (ib *IfBuilder) Then(consequence ast.Expression) *IfBuilder {
	ib.ifExp.Consequence = consequence
	return ib
}

func (ib *IfBuilder) Else(alternative ast.Expression) ast.Expression {
	ib.ifExp.Alternative = alternative
	return ib.ifExp
}

// Case creates a case expression
func (b *Builder) Case(subject ast.Expression) *CaseBuilder {
	return &CaseBuilder{
		caseExp: &ast.CaseExpression{
			Token:   lexer.Token{Type: lexer.CASE},
			Subject: subject,
		},
	}
}

type CaseBuilder struct {
	caseExp *ast.CaseExpression
}

func (cb *CaseBuilder) Branch(variable string, typeName string, expr ast.Expression) *CaseBuilder {
	branch := &ast.CaseBranch{
		Token: lexer.Token{Type: lexer.CASE}, // Add Token
		Variable: &ast.ObjectIdentifier{
			Token: lexer.Token{Type: lexer.OBJECTID, Literal: variable},
			Value: variable,
		},
		Type: &ast.TypeIdentifier{
			Token: lexer.Token{Type: lexer.TYPEID, Literal: typeName},
			Value: typeName,
		},
		Expression: expr,
	}
	cb.caseExp.Branches = append(cb.caseExp.Branches, branch)
	return cb
}

func (cb *CaseBuilder) Build() ast.Expression {
	return cb.caseExp
}

// Let creates a let expression
func (b *Builder) Let() *LetBuilder {
	return &LetBuilder{
		letExp: &ast.LetExpression{
			Token:    lexer.Token{Type: lexer.LET},
			Bindings: []*ast.LetBinding{},
		},
	}
}

type LetBuilder struct {
	letExp *ast.LetExpression
}

func (lb *LetBuilder) Bind(name string, typeName string, init ast.Expression) *LetBuilder {
	binding := &ast.LetBinding{
		Token: lexer.Token{Type: lexer.LET}, // Add Token
		Identifier: &ast.ObjectIdentifier{
			Token: lexer.Token{Type: lexer.OBJECTID, Literal: name},
			Value: name,
		},
		Type: &ast.TypeIdentifier{
			Token: lexer.Token{Type: lexer.TYPEID, Literal: typeName},
			Value: typeName,
		},
		Init: init,
	}
	lb.letExp.Bindings = append(lb.letExp.Bindings, binding)
	return lb
}

func (lb *LetBuilder) In(expr ast.Expression) ast.Expression {
	lb.letExp.In = expr
	return lb.letExp
}

// Call creates a method call expression
func (b *Builder) Call(object ast.Expression, method string) *CallBuilder {
	return &CallBuilder{
		callExp: &ast.MethodCallExpression{
			Token:  lexer.Token{Type: lexer.DOT},
			Object: object,
			Method: &ast.ObjectIdentifier{
				Token: lexer.Token{Type: lexer.OBJECTID, Literal: method},
				Value: method,
			},
		},
	}
}

type CallBuilder struct {
	callExp *ast.MethodCallExpression
}

func (cb *CallBuilder) Args(args ...ast.Expression) ast.Expression {
	cb.callExp.Arguments = args
	return cb.callExp
}

// Block creates a block expression
func (b *Builder) Block() *BlockBuilder {
	return &BlockBuilder{
		block: &ast.BlockExpression{
			Token:       lexer.Token{Type: lexer.LBRACE},
			Expressions: []ast.Expression{},
		},
	}
}

// BlockBuilder builds block expressions
type BlockBuilder struct {
	block *ast.BlockExpression
}

// Add adds an expression to the block
func (bb *BlockBuilder) Add(exp ast.Expression) *BlockBuilder {
	bb.block.Expressions = append(bb.block.Expressions, exp)
	return bb
}

// Build returns the completed block expression
func (bb *BlockBuilder) Build() ast.Expression {
	return bb.block
}

// MethodBuilder builds method declarations
type MethodBuilder struct {
	method *ast.Method
}

// ReturnType sets the method return type
func (mb *MethodBuilder) ReturnType(typeName string) *MethodBuilder {
	mb.method.Type = &ast.TypeIdentifier{
		Token: lexer.Token{Type: lexer.TYPEID, Literal: typeName},
		Value: typeName,
	}
	return mb
}

// Body sets the method body
func (mb *MethodBuilder) Body(body ast.Expression) *MethodBuilder {
	mb.method.Body = body
	return mb
}

// Param adds a parameter to the method
func (mb *MethodBuilder) Param(name string, typeName string) *MethodBuilder {
	formal := &ast.Formal{
		Name: &ast.ObjectIdentifier{
			Token: lexer.Token{Type: lexer.OBJECTID, Literal: name},
			Value: name,
		},
		Type: &ast.TypeIdentifier{
			Token: lexer.Token{Type: lexer.TYPEID, Literal: typeName},
			Value: typeName,
		},
	}
	mb.method.Formals = append(mb.method.Formals, formal)
	return mb
}

// Build returns the completed method
func (mb *MethodBuilder) Build() *ast.Method {
	return mb.method
}
