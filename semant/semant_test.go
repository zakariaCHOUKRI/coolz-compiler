package semant

import (
	"coolz-compiler/ast"
	"coolz-compiler/lexer"
	"coolz-compiler/parser"
	"fmt"
	"strings"
	"testing"
)

func parseProgram(input string) *ast.Program {
	fmt.Println("Starting to parse program")
	l := lexer.NewLexer(strings.NewReader(input))
	p := parser.New(l)
	prog := p.ParseProgram()
	fmt.Printf("Parsed program: %+v\n", prog)
	return prog
}

func TestBasicTypeChecking(t *testing.T) {
	fmt.Println("Starting TestBasicTypeChecking")
	tests := []struct {
		input       string
		shouldError bool
		errorCount  int
	}{
		// Let's start with just one test case to debug
		{
			input: `
				class Main {
					x : Int <- 42;
				};
			`,
			shouldError: false,
		},
	}

	for i, test := range tests {
		fmt.Printf("\nRunning test case %d\n", i)
		fmt.Printf("Input:\n%s\n", test.input)

		program := parseProgram(test.input)
		if program == nil {
			t.Fatalf("Failed to parse program for test %d", i)
		}

		fmt.Printf("Program parsed successfully: %v\n", program)

		analyzer := NewSemanticAnalyser()
		fmt.Println("Created analyzer, starting analysis...")

		analyzer.Analyze(program)
		fmt.Println("Analysis completed")

		hasErrors := len(analyzer.Errors()) > 0
		if hasErrors != test.shouldError {
			t.Errorf("test %d: expected errors = %v, got errors: %v",
				i, test.shouldError, analyzer.Errors())
		}
	}
}

func TestInheritance(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
	}{
		{
			// Valid inheritance
			input: `
				class Parent {
					x : Int <- 42;
				};
				class Child inherits Parent {
					y : Int <- 21;
				};
			`,
			shouldError: false,
		},
		{
			// Invalid inheritance from Int
			input: `
				class Bad inherits Int {
					x : Int <- 42;
				};
			`,
			shouldError: true,
		},
		{
			// Inheritance cycle
			input: `
				class A inherits B {};
				class B inherits A {};
			`,
			shouldError: true,
		},
	}

	for i, test := range tests {
		program := parseProgram(test.input)
		analyzer := NewSemanticAnalyser()
		analyzer.Analyze(program)

		hasErrors := len(analyzer.Errors()) > 0
		if hasErrors != test.shouldError {
			t.Errorf("test %d: expected errors = %v, got errors: %v",
				i, test.shouldError, analyzer.Errors())
		}
	}
}

func TestSELF_TYPE(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
	}{
		{
			// Valid SELF_TYPE usage
			input: `
				class Main {
					self() : SELF_TYPE { self };
					clone() : SELF_TYPE { new SELF_TYPE };
				};
			`,
			shouldError: false,
		},
		{
			// Invalid SELF_TYPE assignment
			input: `
				class Main {
					x : SELF_TYPE <- new Int;
				};
			`,
			shouldError: true,
		},
	}

	for i, test := range tests {
		program := parseProgram(test.input)
		analyzer := NewSemanticAnalyser()
		analyzer.Analyze(program)

		hasErrors := len(analyzer.Errors()) > 0
		if hasErrors != test.shouldError {
			t.Errorf("test %d: expected errors = %v, got errors: %v",
				i, test.shouldError, analyzer.Errors())
		}
	}
}

func TestMethodDispatch(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
	}{
		{
			// Valid method dispatch
			input: `
				class Main {
					foo(x: Int) : Int { x + 1 };
					bar() : Int { foo(42) };
				};
			`,
			shouldError: false,
		},
		{
			// Invalid argument type
			input: `
				class Main {
					foo(x: Int) : Int { x + 1 };
					bar() : Int { foo("string") };
				};
			`,
			shouldError: true,
		},
		{
			// Invalid number of arguments
			input: `
				class Main {
					foo(x: Int) : Int { x + 1 };
					bar() : Int { foo() };
				};
			`,
			shouldError: true,
		},
	}

	for i, test := range tests {
		program := parseProgram(test.input)
		analyzer := NewSemanticAnalyser()
		analyzer.Analyze(program)

		hasErrors := len(analyzer.Errors()) > 0
		if hasErrors != test.shouldError {
			t.Errorf("test %d: expected errors = %v, got errors: %v",
				i, test.shouldError, analyzer.Errors())
		}
	}
}
