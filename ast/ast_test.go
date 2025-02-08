package ast

import (
	"coolz-compiler/lexer"
	"testing"
)

func TestClassNode(t *testing.T) {
	classToken := lexer.Token{Type: lexer.CLASS, Literal: "class"}
	mainType := &TypeIdentifier{
		Token: lexer.Token{Type: lexer.TYPEID, Literal: "Main"},
		Value: "Main",
	}
	class := &Class{
		Token: classToken,
		Name:  mainType,
	}

	if class.TokenLiteral() != "class" {
		t.Errorf("Class.TokenLiteral() = %q, want 'class'", class.TokenLiteral())
	}
}

func TestMethodNode(t *testing.T) {
	methodToken := lexer.Token{Type: lexer.OBJECTID, Literal: "main"}
	returnType := &TypeIdentifier{
		Token: lexer.Token{Type: lexer.TYPEID, Literal: "Object"},
		Value: "Object",
	}
	body := &Block{
		Token: lexer.Token{Type: lexer.LBRACE, Literal: "{"},
	}

	method := &Method{
		Token:      methodToken,
		Name:       &ObjectIdentifier{Token: lexer.Token{Type: lexer.OBJECTID, Literal: "main"}, Value: "main"},
		ReturnType: returnType,
		Body:       body,
	}

	if method.Name.Value != "main" {
		t.Errorf("Method.Name = %q, want 'main'", method.Name.Value)
	}
}

func TestDispatchExpression(t *testing.T) {
	receiver := &SelfExpression{Token: lexer.Token{Type: lexer.SELF}}
	method := &ObjectIdentifier{Token: lexer.Token{Type: lexer.OBJECTID, Literal: "toString"}, Value: "toString"}

	dispatch := &Dispatch{
		Receiver: receiver,
		Method:   method,
	}

	if dispatch.Method.Value != "toString" {
		t.Errorf("Dispatch.Method = %q, want 'toString'", dispatch.Method.Value)
	}
}

func TestBinaryExpression(t *testing.T) {
	left := &IntegerLiteral{Value: 5}
	right := &IntegerLiteral{Value: 3}
	opToken := lexer.Token{Type: lexer.PLUS, Literal: "+"}

	binExpr := &BinaryExpression{
		Left:     left,
		Operator: lexer.PLUS,
		Right:    right,
		Token:    opToken,
	}

	if binExpr.Operator != lexer.PLUS {
		t.Errorf("BinaryExpression.Operator = %v, want PLUS", binExpr.Operator)
	}
}

func TestLetExpression(t *testing.T) {
	letExpr := &Let{
		VarName: &ObjectIdentifier{
			Token: lexer.Token{Type: lexer.OBJECTID, Literal: "x"},
			Value: "x",
		},
		VarType: &TypeIdentifier{
			Token: lexer.Token{Type: lexer.TYPEID, Literal: "Int"},
			Value: "Int",
		},
		VarInit: &IntegerLiteral{Value: 42},
		Body: &ObjectIdentifier{
			Token: lexer.Token{Type: lexer.OBJECTID, Literal: "x"},
			Value: "x",
		},
	}

	if letExpr.VarName.Value != "x" {
		t.Errorf("let.VarName.Value not 'x'. got=%s", letExpr.VarName.Value)
	}

	if letExpr.VarType.Value != "Int" {
		t.Errorf("let.VarType.Value not 'Int'. got=%s", letExpr.VarType.Value)
	}
}

func TestProgramNode(t *testing.T) {
	classToken := lexer.Token{Type: lexer.CLASS, Literal: "class"}
	mainType := &TypeIdentifier{
		Token: lexer.Token{Type: lexer.TYPEID, Literal: "Main"},
		Value: "Main",
	}
	class := &Class{
		Token: classToken,
		Name:  mainType,
	}

	program := &Program{
		Classes: []*Class{class},
	}

	if program.TokenLiteral() != "class" {
		t.Errorf("Program.TokenLiteral() = %q, want 'class'", program.TokenLiteral())
	}
}

func TestAttributeNode(t *testing.T) {
	attrToken := lexer.Token{Type: lexer.OBJECTID, Literal: "x"}
	attrType := &TypeIdentifier{
		Token: lexer.Token{Type: lexer.TYPEID, Literal: "Int"},
		Value: "Int",
	}
	attr := &Attribute{
		Token: attrToken,
		Name:  &ObjectIdentifier{Token: lexer.Token{Type: lexer.OBJECTID, Literal: "x"}, Value: "x"},
		Type:  attrType,
	}

	if attr.Name.Value != "x" {
		t.Errorf("Attribute.Name = %q, want 'x'", attr.Name.Value)
	}
}

func TestConditionalNode(t *testing.T) {
	conditional := &Conditional{
		Token: lexer.Token{Type: lexer.IF, Literal: "if"},
		Predicate: &BooleanLiteral{
			Token: lexer.Token{Type: lexer.BOOL_CONST, Literal: "true"},
			Value: true,
		},
		ThenBranch: &IntegerLiteral{
			Token: lexer.Token{Type: lexer.INT_CONST, Literal: "1"},
			Value: 1,
		},
		ElseBranch: &IntegerLiteral{
			Token: lexer.Token{Type: lexer.INT_CONST, Literal: "0"},
			Value: 0,
		},
	}

	if conditional.Predicate == nil {
		t.Error("conditional.Predicate is nil")
	}

	if conditional.TokenLiteral() != "if" {
		t.Errorf("conditional.TokenLiteral not 'if'. got=%q", conditional.TokenLiteral())
	}
}

func TestLoopNode(t *testing.T) {
	loopToken := lexer.Token{Type: lexer.WHILE, Literal: "while"}
	condition := &BooleanLiteral{Value: true}
	body := &IntegerLiteral{Value: 1}

	loop := &Loop{
		Token:     loopToken,
		Condition: condition,
		Body:      body,
	}

	if loop.Condition != condition {
		t.Errorf("Loop.Condition = %v, want %v", loop.Condition, condition)
	}
}

func TestBlockNode(t *testing.T) {
	blockToken := lexer.Token{Type: lexer.LBRACE, Literal: "{"}
	expr1 := &IntegerLiteral{Value: 1}
	expr2 := &IntegerLiteral{Value: 2}

	block := &Block{
		Token:       blockToken,
		Expressions: []Expression{expr1, expr2},
	}

	if len(block.Expressions) != 2 {
		t.Errorf("Block.Expressions length = %d, want 2", len(block.Expressions))
	}
}

func TestNewNode(t *testing.T) {
	newToken := lexer.Token{Type: lexer.NEW, Literal: "new"}
	newType := &TypeIdentifier{
		Token: lexer.Token{Type: lexer.TYPEID, Literal: "Object"},
		Value: "Object",
	}

	newExpr := &New{
		Token: newToken,
		Type:  newType,
	}

	if newExpr.Type.Value != "Object" {
		t.Errorf("New.Type = %q, want 'Object'", newExpr.Type.Value)
	}
}

func TestIsVoidNode(t *testing.T) {
	isVoidToken := lexer.Token{Type: lexer.ISVOID, Literal: "isvoid"}
	expr := &IntegerLiteral{Value: 42}

	isVoid := &IsVoid{
		Token: isVoidToken,
		Expr:  expr,
	}

	if isVoid.Expr != expr {
		t.Errorf("IsVoid.Expr = %v, want %v", isVoid.Expr, expr)
	}
}

func TestUnaryExpressionNode(t *testing.T) {
	unaryExpr := &UnaryExpression{
		Token:    lexer.Token{Type: lexer.NEG, Literal: "~"},
		Operator: lexer.NEG,
		Right: &IntegerLiteral{
			Token: lexer.Token{Type: lexer.INT_CONST, Literal: "5"},
			Value: 5,
		},
	}

	if unaryExpr.TokenLiteral() != "~" {
		t.Errorf("unaryExpr.TokenLiteral not '~'. got=%q", unaryExpr.TokenLiteral())
	}
}
