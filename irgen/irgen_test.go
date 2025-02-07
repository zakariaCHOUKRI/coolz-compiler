package irgen

import (
	"coolz-compiler/ast"
	"strings"
	"testing"

	"github.com/llir/llvm/ir"
)

func TestIRGenerator_GenerateIntegerLiteral(t *testing.T) {
	program := &ast.Program{
		Classes: []*ast.Class{
			{
				Name: &ast.TypeIdentifier{Value: "Main"},
				Features: []ast.Feature{
					&ast.Method{
						Name: &ast.ObjectIdentifier{Value: "main"},
						Type: &ast.TypeIdentifier{Value: "Int"},
						Body: &ast.IntegerLiteral{Value: 42},
					},
				},
			},
		},
	}

	generator := NewIRGenerator()
	module, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("Failed to generate IR: %v", err)
	}

	// Convert module to string for inspection
	ir := module.String()

	// Basic verification
	if !containsFunction(module, "Int_new") {
		t.Error("Generated IR does not contain Int_new function")
	}

	if !containsValue(ir, "42") {
		t.Error("Generated IR does not contain the integer literal 42")
	}
}

func TestIRGenerator_GenerateStringLiteral(t *testing.T) {
	program := &ast.Program{
		Classes: []*ast.Class{
			{
				Name: &ast.TypeIdentifier{Value: "Main"},
				Features: []ast.Feature{
					&ast.Method{
						Name: &ast.ObjectIdentifier{Value: "main"},
						Type: &ast.TypeIdentifier{Value: "String"},
						Body: &ast.StringLiteral{Value: "Hello, World!"},
					},
				},
			},
		},
	}

	generator := NewIRGenerator()
	module, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("Failed to generate IR: %v", err)
	}

	// Convert module to string for inspection
	ir := module.String()

	// Basic verification
	if !containsFunction(module, "String_new") {
		t.Error("Generated IR does not contain String_new function")
	}

	if !containsValue(ir, "Hello, World!") {
		t.Error("Generated IR does not contain the string literal")
	}
}

// Helper function to check if a function exists in the module
func containsFunction(module *ir.Module, funcName string) bool {
	for _, f := range module.Funcs {
		if f.Name() == funcName {
			return true
		}
	}
	return false
}

// Helper function to check if a string exists in the IR
func containsValue(ir string, value string) bool {
	return strings.Contains(ir, value)
}
