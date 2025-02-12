package parser

import (
	"coolz-compiler/ast"
	"coolz-compiler/lexer"
	"fmt"
	"strings"
	"testing"
)

func TestClassDefinition(t *testing.T) {
	input := `
class Main {
    main() : Object {
        {
            out_string("Hello, World!\n");
        }
    };
};
`

	l := lexer.NewLexer(strings.NewReader(input))
	p := New(l)
	program := p.ParseProgram()

	if len(program.Classes) != 1 {
		t.Fatalf("program.Classes does not contain 1 class. got=%d", len(program.Classes))
	}

	class := program.Classes[0]
	if class.Name.Value != "Main" {
		t.Fatalf("class name not 'Main'. got=%s", class.Name.Value)
	}

	if len(class.Features) != 1 {
		t.Fatalf("class.Features does not contain 1 feature. got=%d", len(class.Features))
	}

	method, ok := class.Features[0].(*ast.Method)
	if !ok {
		t.Fatalf("class.Features[0] is not a method. got=%T", class.Features[0])
	}

	if method.Name.Value != "main" {
		t.Fatalf("method name not 'main'. got=%s", method.Name.Value)
	}

	if method.Type.Value != "Object" {
		t.Fatalf("method return type not 'Object'. got=%s", method.Type.Value)
	}
}

func TestInheritance(t *testing.T) {
	input := `
class Main inherits IO {
    main() : Object {
        {
            out_string("Hello, World!\n");
        }
    };
};
`

	l := lexer.NewLexer(strings.NewReader(input))
	p := New(l)
	program := p.ParseProgram()

	if len(program.Classes) != 1 {
		t.Fatalf("program.Classes does not contain 1 class. got=%d", len(program.Classes))
	}

	class := program.Classes[0]
	if class.Name.Value != "Main" {
		t.Fatalf("class name not 'Main'. got=%s", class.Name.Value)
	}

	if class.Parent.Value != "IO" {
		t.Fatalf("class parent not 'IO'. got=%s", class.Parent.Value)
	}
}

func TestMethodWithFormals(t *testing.T) {
	input := `
class Main {
    add(x : Int, y : Int) : Int {
        x + y
    };
};
`

	l := lexer.NewLexer(strings.NewReader(input))
	p := New(l)
	program := p.ParseProgram()

	if len(program.Classes) != 1 {
		t.Fatalf("program.Classes does not contain 1 class. got=%d", len(program.Classes))
	}

	class := program.Classes[0]
	if class.Name.Value != "Main" {
		t.Fatalf("class name not 'Main'. got=%s", class.Name.Value)
	}

	if len(class.Features) != 1 {
		t.Fatalf("class.Features does not contain 1 feature. got=%d", len(class.Features))
	}

	method, ok := class.Features[0].(*ast.Method)
	if !ok {
		t.Fatalf("class.Features[0] is not a method. got=%T", class.Features[0])
	}

	if method.Name.Value != "add" {
		t.Fatalf("method name not 'add'. got=%s", method.Name.Value)
	}

	if len(method.Formals) != 2 {
		t.Fatalf("method.Formals does not contain 2 formals. got=%d", len(method.Formals))
	}

	if method.Formals[0].Name.Value != "x" {
		t.Fatalf("first formal name not 'x'. got=%s", method.Formals[0].Name.Value)
	}

	if method.Formals[1].Name.Value != "y" {
		t.Fatalf("second formal name not 'y'. got=%s", method.Formals[1].Name.Value)
	}
}

func TestIfExpression(t *testing.T) {
	input := `
class Main {
    main() : Object {
        if true then 1 else 0 fi
    };
};
`

	l := lexer.NewLexer(strings.NewReader(input))
	p := New(l)
	program := p.ParseProgram()

	if len(program.Classes) != 1 {
		t.Fatalf("program.Classes does not contain 1 class. got=%d", len(program.Classes))
	}

	class := program.Classes[0]
	if class.Name.Value != "Main" {
		t.Fatalf("class name not 'Main'. got=%s", class.Name.Value)
	}

	if len(class.Features) != 1 {
		t.Fatalf("class.Features does not contain 1 feature. got=%d", len(class.Features))
	}

	method, ok := class.Features[0].(*ast.Method)
	if !ok {
		t.Fatalf("class.Features[0] is not a method. got=%T", class.Features[0])
	}

	ifExpr, ok := method.Body.(*ast.IfExpression)
	if !ok {
		t.Fatalf("method body is not an if expression. got=%T", method.Body)
	}

	if ifExpr.Condition.(*ast.BooleanLiteral).Value != true {
		t.Fatalf("if condition not true. got=%v", ifExpr.Condition.(*ast.BooleanLiteral).Value)
	}

	if ifExpr.Consequence.(*ast.IntegerLiteral).Value != 1 {
		t.Fatalf("if consequence not 1. got=%d", ifExpr.Consequence.(*ast.IntegerLiteral).Value)
	}

	if ifExpr.Alternative.(*ast.IntegerLiteral).Value != 0 {
		t.Fatalf("if alternative not 0. got=%d", ifExpr.Alternative.(*ast.IntegerLiteral).Value)
	}
}

func TestWhileExpression(t *testing.T) {
	input := `
class Main {
    main() : Object {
        while true loop 1 pool
    };
};
`

	l := lexer.NewLexer(strings.NewReader(input))
	p := New(l)
	program := p.ParseProgram()

	if len(program.Classes) != 1 {
		t.Fatalf("program.Classes does not contain 1 class. got=%d", len(program.Classes))
	}

	class := program.Classes[0]
	if class.Name.Value != "Main" {
		t.Fatalf("class name not 'Main'. got=%s", class.Name.Value)
	}

	if len(class.Features) != 1 {
		t.Fatalf("class.Features does not contain 1 feature. got=%d", len(class.Features))
	}

	method, ok := class.Features[0].(*ast.Method)
	if !ok {
		t.Fatalf("class.Features[0] is not a method. got=%T", class.Features[0])
	}

	whileExpr, ok := method.Body.(*ast.WhileExpression)
	if !ok {
		t.Fatalf("method body is not a while expression. got=%T", method.Body)
	}

	if whileExpr.Condition.(*ast.BooleanLiteral).Value != true {
		t.Fatalf("while condition not true. got=%v", whileExpr.Condition.(*ast.BooleanLiteral).Value)
	}

	if whileExpr.Body.(*ast.IntegerLiteral).Value != 1 {
		t.Fatalf("while body not 1. got=%d", whileExpr.Body.(*ast.IntegerLiteral).Value)
	}
}

func TestLetExpression(t *testing.T) {
	input := `
class Main {
    main() : Object {
        let x : Int <- 1 in x + 2
    };
};
`

	l := lexer.NewLexer(strings.NewReader(input))
	p := New(l)
	program := p.ParseProgram()

	if len(program.Classes) != 1 {
		t.Fatalf("program.Classes does not contain 1 class. got=%d", len(program.Classes))
	}

	class := program.Classes[0]
	if class.Name.Value != "Main" {
		t.Fatalf("class name not 'Main'. got=%s", class.Name.Value)
	}

	if len(class.Features) != 1 {
		t.Fatalf("class.Features does not contain 1 feature. got=%d", len(class.Features))
	}

	method, ok := class.Features[0].(*ast.Method)
	if !ok {
		t.Fatalf("class.Features[0] is not a method. got=%T", class.Features[0])
	}

	letExpr, ok := method.Body.(*ast.LetExpression)
	if !ok {
		t.Fatalf("method body is not a let expression. got=%T", method.Body)
	}

	if len(letExpr.Bindings) != 1 {
		t.Fatalf("let expression does not contain 1 binding. got=%d", len(letExpr.Bindings))
	}

	if letExpr.Bindings[0].Identifier.Value != "x" {
		t.Fatalf("let binding identifier not 'x'. got=%s", letExpr.Bindings[0].Identifier.Value)
	}

	if letExpr.Bindings[0].Type.Value != "Int" {
		t.Fatalf("let binding type not 'Int'. got=%s", letExpr.Bindings[0].Type.Value)
	}

	if letExpr.Bindings[0].Init.(*ast.IntegerLiteral).Value != 1 {
		t.Fatalf("let binding init not 1. got=%d", letExpr.Bindings[0].Init.(*ast.IntegerLiteral).Value)
	}
}

func TestNewExpression(t *testing.T) {
	input := `
class Main {
    main() : Object {
        new Int
    };
};
`

	l := lexer.NewLexer(strings.NewReader(input))
	p := New(l)
	program := p.ParseProgram()

	if len(program.Classes) != 1 {
		t.Fatalf("program.Classes does not contain 1 class. got=%d", len(program.Classes))
	}

	class := program.Classes[0]
	if class.Name.Value != "Main" {
		t.Fatalf("class name not 'Main'. got=%s", class.Name.Value)
	}

	if len(class.Features) != 1 {
		t.Fatalf("class.Features does not contain 1 feature. got=%d", len(class.Features))
	}

	method, ok := class.Features[0].(*ast.Method)
	if !ok {
		t.Fatalf("class.Features[0] is not a method. got=%T", class.Features[0])
	}

	newExpr, ok := method.Body.(*ast.NewExpression)
	if !ok {
		t.Fatalf("method body is not a new expression. got=%T", method.Body)
	}

	if newExpr.Type.Value != "Int" {
		t.Fatalf("new expression type not 'Int'. got=%s", newExpr.Type.Value)
	}
}

func TestIsVoidExpression(t *testing.T) {
	input := `
class Main {
    main() : Object {
        isvoid 1
    };
};
`

	l := lexer.NewLexer(strings.NewReader(input))
	p := New(l)
	program := p.ParseProgram()

	if len(program.Classes) != 1 {
		t.Fatalf("program.Classes does not contain 1 class. got=%d", len(program.Classes))
	}

	class := program.Classes[0]
	if class.Name.Value != "Main" {
		t.Fatalf("class name not 'Main'. got=%s", class.Name.Value)
	}

	if len(class.Features) != 1 {
		t.Fatalf("class.Features does not contain 1 feature. got=%d", len(class.Features))
	}

	method, ok := class.Features[0].(*ast.Method)
	if !ok {
		t.Fatalf("class.Features[0] is not a method. got=%T", class.Features[0])
	}

	isVoidExpr, ok := method.Body.(*ast.IsVoidExpression)
	if !ok {
		t.Fatalf("method body is not an isvoid expression. got=%T", method.Body)
	}

	if isVoidExpr.Expression.(*ast.IntegerLiteral).Value != 1 {
		t.Fatalf("isvoid expression not 1. got=%d", isVoidExpr.Expression.(*ast.IntegerLiteral).Value)
	}
}

func TestCaseExpression(t *testing.T) {
	input := `
class Main {
    main() : Object {
        case 1 of
            x : Int => x + 1;
            y : String => y;
        esac
    };
};
`

	l := lexer.NewLexer(strings.NewReader(input))
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Errorf("parser has %d errors:", len(p.Errors()))
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
	}

	// Add debugging output
	method := program.Classes[0].Features[0].(*ast.Method)
	t.Logf("Method body type: %T", method.Body)
	if caseExpr, ok := method.Body.(*ast.CaseExpression); ok {
		t.Logf("Case expression details: Expr=%T, Branches=%d",
			caseExpr.Expr, len(caseExpr.Branches))
	}

	if len(program.Classes) != 1 {
		t.Fatalf("program.Classes does not contain 1 class. got=%d", len(program.Classes))
	}

	class := program.Classes[0]
	if class.Name.Value != "Main" {
		t.Fatalf("class name not 'Main'. got=%s", class.Name.Value)
	}

	if len(class.Features) != 1 {
		t.Fatalf("class.Features does not contain 1 feature. got=%d", len(class.Features))
	}

	method, ok := class.Features[0].(*ast.Method)
	if !ok {
		t.Fatalf("class.Features[0] is not a method. got=%T", class.Features[0])
	}

	caseExpr, ok := method.Body.(*ast.CaseExpression)
	if !ok {
		t.Fatalf("method body is not a case expression. got=%T", method.Body)
	}

	if len(caseExpr.Branches) != 2 {
		t.Fatalf("case expression does not contain 2 branches. got=%d", len(caseExpr.Branches))
	}

	if caseExpr.Branches[0].Identifier.Value != "x" {
		t.Fatalf("first branch identifier not 'x'. got=%s", caseExpr.Branches[0].Identifier.Value)
	}

	if caseExpr.Branches[0].Type.Value != "Int" {
		t.Fatalf("first branch type not 'Int'. got=%s", caseExpr.Branches[0].Type.Value)
	}

	if caseExpr.Branches[1].Identifier.Value != "y" {
		t.Fatalf("second branch identifier not 'y'. got=%s", caseExpr.Branches[1].Identifier.Value)
	}

	if caseExpr.Branches[1].Type.Value != "String" {
		t.Fatalf("second branch type not 'String'. got=%s", caseExpr.Branches[1].Type.Value)
	}
}

func TestDynamicDispatch(t *testing.T) {
	input := `
class Main {
    main() : Object {
        self.foo(1, 2)
    };
};
`

	l := lexer.NewLexer(strings.NewReader(input))
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Errorf("parser has %d errors:", len(p.Errors()))
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
	}

	// Add debugging output
	method := program.Classes[0].Features[0].(*ast.Method)
	t.Logf("Method body type: %T", method.Body)
	if dispatch, ok := method.Body.(*ast.DynamicDispatch); ok {
		t.Logf("Dispatch details: Object=%T, Method=%s, Args=%d",
			dispatch.Object, dispatch.Method.Value, len(dispatch.Arguments))
	}

	if len(program.Classes) != 1 {
		t.Fatalf("program.Classes does not contain 1 class. got=%d", len(program.Classes))
	}

	class := program.Classes[0]
	if class.Name.Value != "Main" {
		t.Fatalf("class name not 'Main'. got=%s", class.Name.Value)
	}

	if len(class.Features) != 1 {
		t.Fatalf("class.Features does not contain 1 feature. got=%d", len(class.Features))
	}

	method, ok := class.Features[0].(*ast.Method)
	if !ok {
		t.Fatalf("class.Features[0] is not a method. got=%T", class.Features[0])
	}

	dispatchExpr, ok := method.Body.(*ast.DynamicDispatch)
	if !ok {
		t.Fatalf("method body is not a dynamic dispatch expression. got=%T", method.Body)
	}

	if dispatchExpr.Object.(*ast.Self).TokenLiteral() != "self" {
		t.Fatalf("dispatch object not 'self'. got=%s", dispatchExpr.Object.(*ast.Self).TokenLiteral())
	}

	if dispatchExpr.Method.Value != "foo" {
		t.Fatalf("dispatch method not 'foo'. got=%s", dispatchExpr.Method.Value)
	}

	if len(dispatchExpr.Arguments) != 2 {
		t.Fatalf("dispatch does not contain 2 arguments. got=%d", len(dispatchExpr.Arguments))
	}

	if dispatchExpr.Arguments[0].(*ast.IntegerLiteral).Value != 1 {
		t.Fatalf("first argument not 1. got=%d", dispatchExpr.Arguments[0].(*ast.IntegerLiteral).Value)
	}

	if dispatchExpr.Arguments[1].(*ast.IntegerLiteral).Value != 2 {
		t.Fatalf("second argument not 2. got=%d", dispatchExpr.Arguments[1].(*ast.IntegerLiteral).Value)
	}
}

func TestStaticDispatch(t *testing.T) {
	input := `
class Main {
    main() : Object {
        self@IO.out_string("Hello, World!\n")
    };
};
`

	l := lexer.NewLexer(strings.NewReader(input))
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Errorf("parser has %d errors:", len(p.Errors()))
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
	}

	// Add debugging output
	method := program.Classes[0].Features[0].(*ast.Method)
	t.Logf("Method body type: %T", method.Body)
	if dispatch, ok := method.Body.(*ast.StaticDispatch); ok {
		t.Logf("Static dispatch details: Object=%T, Type=%s, Method=%s, Args=%d",
			dispatch.Object, dispatch.Type.Value, dispatch.Method.Value, len(dispatch.Arguments))
	}

	if len(program.Classes) != 1 {
		t.Fatalf("program.Classes does not contain 1 class. got=%d", len(program.Classes))
	}

	class := program.Classes[0]
	if class.Name.Value != "Main" {
		t.Fatalf("class name not 'Main'. got=%s", class.Name.Value)
	}

	if len(class.Features) != 1 {
		t.Fatalf("class.Features does not contain 1 feature. got=%d", len(class.Features))
	}

	method, ok := class.Features[0].(*ast.Method)
	if !ok {
		t.Fatalf("class.Features[0] is not a method. got=%T", class.Features[0])
	}

	dispatchExpr, ok := method.Body.(*ast.StaticDispatch)
	if !ok {
		t.Fatalf("method body is not a static dispatch expression. got=%T", method.Body)
	}

	if dispatchExpr.Object.(*ast.Self).TokenLiteral() != "self" {
		t.Fatalf("dispatch object not 'self'. got=%s", dispatchExpr.Object.(*ast.Self).TokenLiteral())
	}

	if dispatchExpr.Type.Value != "IO" {
		t.Fatalf("static dispatch type not 'IO'. got=%s", dispatchExpr.Type.Value)
	}

	if dispatchExpr.Method.Value != "out_string" {
		t.Fatalf("dispatch method not 'out_string'. got=%s", dispatchExpr.Method.Value)
	}

	if len(dispatchExpr.Arguments) != 1 {
		t.Fatalf("dispatch does not contain 1 argument. got=%d", len(dispatchExpr.Arguments))
	}

	strArg, ok := dispatchExpr.Arguments[0].(*ast.StringLiteral)
	if !ok {
		t.Fatalf("dispatch argument is not a string literal. got=%T", dispatchExpr.Arguments[0])
	}

	if strArg.Value != "Hello, World!\n" {
		t.Fatalf("dispatch argument not 'Hello, World!\n'. got=%s", strArg.Value)
	}
}

func TestMethodCallWithMultipleArguments(t *testing.T) {
	input := `
class Main {
    main() : Object {
        foo(1, "hello", true, 42)
    };
};
`

	l := lexer.NewLexer(strings.NewReader(input))
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Errorf("parser has %d errors:", len(p.Errors()))
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
	}

	if len(program.Classes) != 1 {
		t.Fatalf("program.Classes does not contain 1 class. got=%d", len(program.Classes))
	}

	class := program.Classes[0]
	method, ok := class.Features[0].(*ast.Method)
	if !ok {
		t.Fatalf("class.Features[0] is not a method. got=%T", class.Features[0])
	}

	// The method body should be a dynamic dispatch
	dispatch, ok := method.Body.(*ast.DynamicDispatch)
	if !ok {
		t.Fatalf("method body is not a dynamic dispatch. got=%T", method.Body)
	}

	// Check method name
	if dispatch.Method.Value != "foo" {
		t.Errorf("method name not 'foo'. got=%s", dispatch.Method.Value)
	}

	// Check number of arguments
	if len(dispatch.Arguments) != 4 {
		t.Fatalf("wrong number of arguments. expected=4, got=%d", len(dispatch.Arguments))
	}

	// Check first argument (integer)
	intArg, ok := dispatch.Arguments[0].(*ast.IntegerLiteral)
	if !ok || intArg.Value != 1 {
		t.Errorf("first argument not int 1. got=%T(%+v)", dispatch.Arguments[0], dispatch.Arguments[0])
	}

	// Check second argument (string)
	strArg, ok := dispatch.Arguments[1].(*ast.StringLiteral)
	if !ok || strArg.Value != "hello" {
		t.Errorf("second argument not string 'hello'. got=%T(%+v)", dispatch.Arguments[1], dispatch.Arguments[1])
	}

	// Check third argument (boolean)
	boolArg, ok := dispatch.Arguments[2].(*ast.BooleanLiteral)
	if !ok || !boolArg.Value {
		t.Errorf("third argument not boolean true. got=%T(%+v)", dispatch.Arguments[2], dispatch.Arguments[2])
	}

	// Check fourth argument (integer)
	intArg2, ok := dispatch.Arguments[3].(*ast.IntegerLiteral)
	if !ok || intArg2.Value != 42 {
		t.Errorf("fourth argument not int 42. got=%T(%+v)", dispatch.Arguments[3], dispatch.Arguments[3])
	}
}

func TestPrattParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"class Main { main(): Object { 1 + 2 * 3 }; };",
			"(1 + (2 * 3))",
		},
		{
			"class Main { main(): Object { 4 * 2 + 3 }; };",
			"((4 * 2) + 3)",
		},
		{
			"class Main { main(): Object { 1 + 2 + 3 }; };",
			"((1 + 2) + 3)",
		},
		{
			"class Main { main(): Object { 1 * 2 * 3 }; };",
			"((1 * 2) * 3)",
		},
		{
			"class Main { main(): Object { 1 + 2 * 3 + 4 }; };",
			"((1 + (2 * 3)) + 4)",
		},
		{
			"class Main { main(): Object { (1 + 2) * 3 }; };",
			"((1 + 2) * 3)",
		},
	}

	for i, tt := range tests {
		l := lexer.NewLexer(strings.NewReader(tt.input))
		p := New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Errorf("test[%d] parser has %d errors:", i, len(p.Errors()))
			for _, err := range p.Errors() {
				t.Errorf("parser error: %s", err)
			}
			continue
		}

		method := program.Classes[0].Features[0].(*ast.Method)
		expression := method.Body

		got := inorderTraversal(expression)
		if got != tt.expected {
			t.Errorf("test[%d] wrong expression.\nexpected=%q\ngot=%q",
				i, tt.expected, got)
		}
	}
}

// Helper function to traverse the AST and generate a string representation
func inorderTraversal(node ast.Expression) string {
	switch n := node.(type) {
	case *ast.BinaryExpression:
		return fmt.Sprintf("(%s %s %s)",
			inorderTraversal(n.Left),
			n.Operator,
			inorderTraversal(n.Right))
	case *ast.IntegerLiteral:
		return fmt.Sprintf("%d", n.Value)
	default:
		return fmt.Sprintf("unknown(%T)", n)
	}
}
