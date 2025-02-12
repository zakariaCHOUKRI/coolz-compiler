package semant

import (
	"coolz-compiler/lexer"
	"coolz-compiler/parser"
	"strings"
	"testing"
)

func parseProgram(input string) *parser.Parser {
	l := lexer.NewLexer(strings.NewReader(input))
	p := parser.New(l)
	p.ParseProgram() // Ignore parse errors for test setup
	return p
}

func TestSemanticAnalysis(t *testing.T) {
	tests := []struct {
		name     string
		program  string
		expected []string
		notes    string
	}{
		{
			name: "Basic Class Structure",
			program: `
				class Main {
					main() : Object {
						42
					};
				};
			`,
			expected: []string{},
		},
		{
			name: "Class Redefinition",
			program: `
				class A {};
				class A {};
			`,
			expected: []string{"class A redefined"},
		},
		{
			name: "Inheritance from Undefined Class",
			program: `
				class B inherits A {};
			`,
			expected: []string{"class A is not defined"},
		},
		{
			name: "Cyclic Inheritance",
			program: `
				class A inherits B {};
				class B inherits A {};
			`,
			expected: []string{"cyclic inheritance detected"},
		},
		{
			name: "Main Class Requirements",
			program: `
				class Main {
					-- Missing main method
				};
			`,
			expected: []string{"Main class must have method main() : Object"},
		},
		{
			name: "Attribute Scope",
			program: `
				class A {
					a : Int;
					a : String;  -- Redefinition
				};
			`,
			expected: []string{"attribute a redefined"},
		},
		{
			name: "Method Overriding",
			program: `
				class A { m() : Int { 0 } };
				class B inherits A { m() : Int { 1 } };  -- Valid override
			`,
			expected: []string{},
		},
		{
			name: "Invalid Method Override",
			program: `
				class A { m() : Int { 0 } };
				class B inherits A { m() : String { "0" } };
			`,
			expected: []string{"method m has incompatible return type"},
		},
		{
			name: "Type Conformance",
			program: `
				class A {};
				class B inherits A {
					test() : A {
						new B  -- B conforms to A
					};
				};
			`,
			expected: []string{},
		},
		{
			name: "Case Expression Type Join",
			program: `
				class Main {
					main() : Object {
						case 42 of
							x : Int => "string";
							y : Bool => true;
						esac
					};
				};
			`,
			expected: []string{"case expression type join: Object"},
			notes:    "TODO: Implement type join calculation",
		},
		{
			name: "SELF_TYPE Handling",
			program: `
				class A {
					copy() : SELF_TYPE { self };
				};
			`,
			expected: []string{},
		},
		{
			name: "Let Expression Scoping",
			program: `
				class Main {
					main() : Object {
						let x : Int <- "string" in x  -- Type mismatch
					};
				};
			`,
			expected: []string{"type String does not conform to Int"},
		},
		{
			name: "Static Dispatch Validation",
			program: `
				class A {
					m() : Int { 0 };
				};
				class B {
					test() : Int {
						(new A)@B.m()  -- Invalid static dispatch
					};
				};
			`,
			expected: []string{"type A does not conform to B"},
		},
		{
			name: "Basic Operation Types",
			program: `
				class Main {
					main() : Object {
						"hello" + "world"  -- Invalid string addition
					};
				};
			`,
			expected: []string{"arithmetic operation on non-Int types"},
		},
		{
			name: "Void Initialization",
			program: `
				class A {
					x : B;  -- B is undefined, defaults to Object
				};
			`,
			expected: []string{"undefined type B"},
		},
		{
			name: "Method Parameter Scope",
			program: `
				class A {
					m(x : Int, x : String) : Int { 0 };  -- Duplicate parameter
				};
			`,
			expected: []string{"duplicate parameter x"},
		},
		{
			name: "Self Assignment",
			program: `
				class A {
					test() : Object {
						self <- new A  -- Invalid self assignment
					};
				};
			`,
			expected: []string{"cannot assign to self"},
		},
		{
			name: "New Expression Validation",
			program: `
				class Main {
					main() : Object {
						new UndefinedType
					};
				};
			`,
			expected: []string{"undefined type UndefinedType"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parseProgram(tt.program)
			sa := NewSemanticAnalyser()
			sa.Analyze(p.ParseProgram())

			// Check expected errors
			if len(sa.errors) != len(tt.expected) {
				t.Errorf("Expected %d errors, got %d", len(tt.expected), len(sa.errors))
				for _, err := range sa.errors {
					t.Logf("Error: %s", err)
				}
				return
			}

			for i, expectedErr := range tt.expected {
				if !strings.Contains(sa.errors[i], expectedErr) {
					t.Errorf("Error %d:\nExpected: %s\nGot: %s", i, expectedErr, sa.errors[i])
				}
			}

			if tt.notes != "" {
				t.Logf("Note: %s", tt.notes)
			}
		})
	}
}

func TestPositiveCases(t *testing.T) {
	validPrograms := []struct {
		name    string
		program string
	}{
		{
			name: "Complete Program",
			program: `
				class Main {
					main() : Object {
						(let x : Int <- 42 in
							case x of
								y : Int => y + 1;
								z : Object => abort();
							esac
						)
					};
				};
				
				class A {
					method(x : Int, y : String) : Bool {
						x = y.length()
					};
				};
			`,
		},
		{
			name: "Inheritance Chain",
			program: `
				class A {};
				class B inherits A {};
				class C inherits B {};
			`,
		},
		{
			name: "Method Overriding with SELF_TYPE",
			program: `
				class A {
					copy() : SELF_TYPE { self };
				};
				class B inherits A {
					copy() : SELF_TYPE { self };
				};
			`,
		},
	}

	for _, tt := range validPrograms {
		t.Run(tt.name, func(t *testing.T) {
			p := parseProgram(tt.program)
			sa := NewSemanticAnalyser()
			sa.Analyze(p.ParseProgram())

			if len(sa.errors) > 0 {
				t.Errorf("Expected no errors, got %d:", len(sa.errors))
				for _, err := range sa.errors {
					t.Errorf("  %s", err)
				}
			}
		})
	}
}
