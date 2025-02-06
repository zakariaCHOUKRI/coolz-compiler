package parser

import (
	"coolz-compiler/ast"
	"coolz-compiler/lexer"
	"strings"
	"testing"
)

func newParserFromInput(input string) *Parser {
	l := lexer.NewLexer(strings.NewReader(input))
	return New(l)
}

func checkParserErrors(t *testing.T, p *Parser, i int) {
	errors := p.Errors()
	if len(errors) > 0 {
		t.Errorf("parser has %d errors for test case %d", len(errors), i)
		for _, msg := range errors {
			t.Errorf("parser error: %q", msg)
		}
		t.FailNow()
	}
}

func TestClassParser(t *testing.T) {
	tests := []struct {
		input          string
		expectedName   string
		expectedParent string
	}{
		{
			input:          "class Main { };",
			expectedName:   "Main",
			expectedParent: "",
		},
		{
			input:          "class A { age:Integer <- 30; };",
			expectedName:   "A",
			expectedParent: "",
		},
		{
			input:          "class B { func(): Void { }; };",
			expectedName:   "B",
			expectedParent: "",
		},
		{
			input:          "class B inherits A { func(): Void { }; };",
			expectedName:   "B",
			expectedParent: "A",
		},
	}

	for i, tt := range tests {
		parser := newParserFromInput(tt.input)
		class := parser.parseClass()

		checkParserErrors(t, parser, i)

		if class.Name.Value != tt.expectedName {
			t.Fatalf("[%q]: expected class name to be %q got %q", tt.input, tt.expectedName, class.Name.Value)
		}

		if class.Parent != nil {
			if class.Parent.Value != tt.expectedParent {
				t.Fatalf("[%q]: expected class parent to be %q got %q", tt.input, tt.expectedParent, class.Parent.Value)
			}
		} else if tt.expectedParent != "" {
			t.Fatalf("[%q]: expected class parent to be %q got nil", tt.input, tt.expectedParent)
		}
	}
}

func TestFormalParsing(t *testing.T) {
	tests := []struct {
		input         string
		expectedNames []string
		expectedTypes []string
	}{
		{
			input:         "var1:Integer",
			expectedNames: []string{"var1"},
			expectedTypes: []string{"Integer"},
		},
		{
			input:         "var1:Integer,var2:Boolean,var3:String",
			expectedNames: []string{"var1", "var2", "var3"},
			expectedTypes: []string{"Integer", "Boolean", "String"},
		},
	}

	for _, tt := range tests {
		parser := newParserFromInput(tt.input)
		formals := parser.parseFormals()

		if len(parser.errors) > 0 {
			for _, err := range parser.errors {
				t.Errorf("Parsing Error %s\n", err)
			}
			t.Fatalf("[%q]: Found errors while parsing", tt.input)
		}

		if len(formals) != len(tt.expectedNames) {
			t.Fatalf("[%q]: expected %d formals got %d", tt.input, len(tt.expectedNames), len(formals))
		}

		for i, formal := range formals {
			if formal.Name.Value != tt.expectedNames[i] {
				t.Fatalf("[%q]: expected formal name to be %q got %q", tt.input, tt.expectedNames[i], formal.Name.Value)
			}
			if formal.Type.Value != tt.expectedTypes[i] {
				t.Fatalf("[%q]: expected formal type to be %q got %q", tt.input, tt.expectedTypes[i], formal.Type.Value)
			}
		}
	}
}

func TestMethodParsing(t *testing.T) {
	tests := []struct {
		input               string
		expectedMethodName  string
		expectedFormalNames []string
		expectedFormalTypes []string
		expectedMethodType  string
	}{
		{
			input:               "main(): Void { 1 }",
			expectedMethodName:  "main",
			expectedFormalNames: []string{},
			expectedFormalTypes: []string{},
			expectedMethodType:  "Void",
		},
		{
			input:               "sum(a:Integer,b:Integer): Integer { 1 }",
			expectedMethodName:  "sum",
			expectedFormalNames: []string{"a", "b"},
			expectedFormalTypes: []string{"Integer", "Integer"},
			expectedMethodType:  "Integer",
		},
	}

	for i, tt := range tests {
		parser := newParserFromInput(tt.input)
		feature := parser.parseFeature()

		if len(parser.Errors()) > 0 {
			t.Errorf("parser has errors: %v", parser.Errors())
			continue
		}

		method, ok := feature.(*ast.Method)
		if !ok {
			t.Fatalf("Expected *ast.Method, got %T", feature)
		}

		checkParserErrors(t, parser, i)

		if method.Name.Value != tt.expectedMethodName {
			t.Fatalf("[%q]: Expected method name to be %q found %q", tt.input, tt.expectedMethodName, method.Name.Value)
		}

		for i, formal := range method.Formals {
			if formal.Name.Value != tt.expectedFormalNames[i] {
				t.Fatalf("[%q]: Expected formal name to be %q found %q", tt.input, tt.expectedFormalNames[i], formal.Name.Value)
			}
			if formal.Type.Value != tt.expectedFormalTypes[i] {
				t.Fatalf("[%q]: Expected formal type to be %q found %q", tt.input, tt.expectedFormalTypes[i], formal.Type.Value)
			}
		}

		if method.Type.Value != tt.expectedMethodType {
			t.Fatalf("[%q]: Expected method type to be %q found %q", tt.input, tt.expectedMethodType, method.Type.Value)
		}
	}
}

func TestAttributeParsing(t *testing.T) {
	tests := []struct {
		input        string
		expectedName string
		expectedType string
	}{
		{
			input:        "firstName:String;",
			expectedName: "firstName",
			expectedType: "String",
		},
		{
			input:        "age:Integer <- 0;",
			expectedName: "age",
			expectedType: "Integer",
		},
	}

	for i, tt := range tests {
		parser := newParserFromInput(tt.input)
		feature := parser.parseFeature()

		checkParserErrors(t, parser, i)

		attribute, ok := feature.(*ast.Attribute)
		if !ok {
			t.Fatalf("[%q]: Expected *ast.Attribute, got %T", tt.input, feature)
		}

		if attribute.Name.Value != tt.expectedName {
			t.Fatalf("[%q]: Expected attribute name to be %q got %q", tt.input, tt.expectedName, attribute.Name.Value)
		}
		if attribute.Type.Value != tt.expectedType {
			t.Fatalf("[%q]: Expected attribute type to be %q got %q", tt.input, tt.expectedType, attribute.Type.Value)
		}
	}
}

func TestExpressionParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"5", "5"},
		{`"hello world"`, `"hello world"`},
		{"true", "true"},
		{"false", "false"},
		{"x", "x"},
		{"not true", "(not true)"},
		{"1 + 2", "(1 + 2)"},
		{"1 < 2", "(1 < 2)"},
		{"1 <= 2", "(1 <= 2)"},
		{"~1", "(~ 1)"}, // 9
		{"1 = 2", "(1 = 2)"},
		{"1 * 2", "(1 * 2)"},
		{"isvoid 1", "isvoid 1"},
		{"1 / 2", "(1 / 2)"},
		// TODO: Implement parenthesis parsing
		{"(1 + 2)", "(1 + 2)"},
		{"new Object", "new Object"},
		{"x <- 5", "(x <- 5)"},                                   // 15
		{"if true then 1 else 2 fi", "if true then 1 else 2 fi"}, // 16
		{"while true loop 1 pool", "while true loop 1 pool"},     // 17
		{"1 + 2 * 3", "(1 + (2 * 3))"},
		{"1 * 2 + 3", "((1 * 2) + 3)"},
		{"x.foo()", "((x . foo))"},
		{"x.foo(1,2)", "((x . foo))"}, // arguments not printed in this example
		{"self", "self"},
		{"foo()", "foo()"},
		{"foo(1, 2)", "foo(1, 2)"},
		{"obj.method(1, 2)", "((obj . method))"},
		{"let x:Int <- 5 in x + 1", "let x:Int <- 5 in (x + 1)"},
	}

	for i, tt := range tests {
		p := newParserFromInput(tt.input)
		expression := p.parseExpression(LOWEST)

		if expression == nil {
			t.Errorf("test [%d] failed to parse expression: %q", i, tt.input)
			continue
		}

		actual := SerializeExpression(expression) // This will now use the function from serializer.go

		if actual != tt.expected {
			t.Errorf("test [%d] expected expression to be '%s', got '%s'", i, tt.expected, actual)
		}
	}
}

func TestMethodBodyParsing(t *testing.T) {
	input := `method(): Int { 1 + 2 }`
	parser := newParserFromInput(input)
	feature := parser.parseFeature()
	method, ok := feature.(*ast.Method)
	if !ok {
		t.Fatalf("Expected *ast.Method, got %T", feature)
	}

	checkParserErrors(t, parser, 0)

	if method.Body == nil {
		t.Fatal("Method body is nil")
	}

	binaryExp, ok := method.Body.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("method.Body is not ast.BinaryExpression. got=%T", method.Body)
	}

	if binaryExp.Operator != "+" {
		t.Errorf("binaryExp.Operator is not '+'. got=%q", binaryExp.Operator)
	}
}

func TestBlockExpressionParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"{ 1; 2; 3; }",
			"{ 1; 2; 3 }",
		},
		{
			"{ x <- 1; y <- 2; x + y; }",
			"{ (x <- 1); (y <- 2); (x + y) }",
		},
	}

	for i, tt := range tests {
		p := newParserFromInput(tt.input)
		expression := p.parseExpression(LOWEST)

		if expression == nil {
			t.Errorf("test [%d] failed to parse expression: %q", i, tt.input)
			continue
		}

		actual := SerializeExpression(expression)
		if actual != tt.expected {
			t.Errorf("test [%d] expected expression to be '%s', got '%s'", i, tt.expected, actual)
		}
	}
}

func TestCaseExpressionParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"case x of a:Int => 1; b:String => 2; c:Bool => 3; esac",
			"case x of a : Int => 1; b : String => 2; c : Bool => 3 esac",
		},
		{
			"case y of i:Int => i + 1; s:String => s.length(); esac",
			"case y of i : Int => (i + 1); s : String => ((s . length)) esac",
		},
	}

	for i, tt := range tests {
		p := newParserFromInput(tt.input)
		expression := p.parseExpression(LOWEST)

		if expression == nil {
			t.Errorf("test [%d] failed to parse expression: %q", i, tt.input)
			continue
		}

		_, ok := expression.(*ast.CaseExpression)
		if !ok {
			t.Fatalf("test [%d] expression is not ast.CaseExpression. got=%T", i, expression)
		}

		actual := SerializeExpression(expression)
		if actual != tt.expected {
			t.Errorf("test [%d] expected expression to be '%s', got '%s'", i, tt.expected, actual)
		}
	}
}

func TestMethodCallWithArguments(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"obj.method(1, 2, 3)",
			"((obj . method))",
		},
		{
			"method(1, true, \"hello\")",
			`method(1, true, "hello")`,
		},
	}

	for i, tt := range tests {
		p := newParserFromInput(tt.input)
		expression := p.parseExpression(LOWEST)

		if expression == nil {
			t.Errorf("test [%d] failed to parse expression: %q", i, tt.input)
			continue
		}

		actual := SerializeExpression(expression)
		if actual != tt.expected {
			t.Errorf("test [%d] expected expression to be '%s', got '%s'", i, tt.expected, actual)
		}
	}
}

func TestStringEscaping(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello\n"`, `"hello\n"`},
		{`"tab\there"`, `"tab\there"`},
		{`"quotes\"here"`, `"quotes\"here"`},
	}

	for i, tt := range tests {
		p := newParserFromInput(tt.input)
		expression := p.parseExpression(LOWEST)

		if expression == nil {
			t.Errorf("test [%d] failed to parse expression: %q", i, tt.input)
			continue
		}

		actual := SerializeExpression(expression)
		if actual != tt.expected {
			t.Errorf("test [%d] expected expression to be '%s', got '%s'", i, tt.expected, actual)
		}
	}
}

func TestPrattParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Start with simpler cases first
		{"1", "1"},
		{"1 + 2", "(1 + 2)"},
		// Basic operators
		{"1 + 2 + 3", "((1 + 2) + 3)"},
		{"1 + 2 * 3", "(1 + (2 * 3))"},
		{"1 * 2 * 3", "((1 * 2) * 3)"},
		{"1 * 2 + 3 * 4", "((1 * 2) + (3 * 4))"},
		{"1 + 2 * 3 + 4", "((1 + (2 * 3)) + 4)"},

		// Comparison operators
		{"1 < 2 = true", "((1 < 2) = true)"},
		{"not 1 < 2", "(not (1 < 2))"},
		{"1 <= 2 * 3", "(1 <= (2 * 3))"},

		// Nested expressions
		{"(1 + 2) * 3", "((1 + 2) * 3)"},
		{"(1 * 2) + (3 * 4)", "((1 * 2) + (3 * 4))"},

		// Method calls and dot expressions
		{"obj.method(1 + 2, 3 * 4)", "((obj . method))"},
		{"a <- b.method()", "(a <- ((b . method)))"},

		// Complex expressions
		{"let x:Int <- 1 + 2 * 3 in x + 4", "let x:Int <- (1 + (2 * 3)) in (x + 4)"},
		{"if 1 < 2 * 3 then 4 + 5 else 6 * 7 fi", "if (1 < (2 * 3)) then (4 + 5) else (6 * 7) fi"},
		{"while 1 + 2 <= 3 * 4 loop 5 + 6 pool", "while ((1 + 2) <= (3 * 4)) loop (5 + 6) pool"},
	}

	for i, tt := range tests {
		p := newParserFromInput(tt.input)
		expression := p.parseExpression(LOWEST)

		if expression == nil {
			t.Errorf("test [%d] failed to parse expression: %q", i, tt.input)
			continue
		}

		actual := SerializeExpression(expression)
		if actual != tt.expected {
			t.Errorf("test [%d] expected expression to be '%s', got '%s'", i, tt.expected, actual)
		}
	}
}
