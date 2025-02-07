package semant_test

import (
	"coolz-compiler/lexer"
	"coolz-compiler/parser"
	"coolz-compiler/semant"
	"strings"
	"testing"
)

func TestSemanticAnalyser(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name: "Valid class with simple attribute initialization",
			input: `
			class Main {
				x : Int <- 42;
			};
			`,
			wantError: false,
		},
		{
			name: "Class with mismatched types in attribute",
			input: `
			class Main {
				x : String <- 42;
			};
			`,
			wantError: true,
		},
		{
			name: "If expression with ambiguous then/else return",
			input: `
			class Main {
				x : Int <- if true then 42 else "Hello" fi;
			};
			`,
			wantError: true,
		},
		// {
		// 	name: "While expression with non-bool condition",
		// 	input: `
		// 	class Main {
		// 		foo() : Int {
		// 			while 100 loop
		// 				"do something";
		// 			pool
		// 		};
		// 	};
		// 	`,
		// 	wantError: true,
		// },
		{
			name: "Simple inheritance test",
			input: `
			class A { };
			class B inherits A { };
			`,
			wantError: false,
		},
		{
			name: "Undefined variable in assignment",
			input: `
			class Main {
				x : Int;
				foo() : Int {
					y <- x;
				};
			};
			`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(strings.NewReader(tt.input))
			p := parser.New(l)
			prog := p.ParseProgram()
			if len(p.Errors()) > 0 {
				if !tt.wantError {
					t.Fatalf("parser error: %v", p.Errors())
				}
				return
			}
			sa := semant.NewSemanticAnalyser()
			sa.Analyze(prog)
			errs := sa.Errors()
			if (len(errs) > 0) != tt.wantError {
				t.Fatalf("expected wantError=%v, got errors=%v", tt.wantError, errs)
			}
		})
	}
}
