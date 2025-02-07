package main

import (
	"coolz-compiler/ast"
	"coolz-compiler/codegen"
	"coolz-compiler/irgen"
	"coolz-compiler/optimizer"
	"fmt"
	"os"
)

func main() {
	// Create AST
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

	// Generate LLVM IR
	generator := irgen.NewIRGenerator()
	module, err := generator.Generate(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating IR: %v\n", err)
		os.Exit(1)
	}

	// Optimize the IR
	opt := optimizer.NewOptimizer()
	opt.SetLevel(optimizer.MediumOptimization) // Access constant directly from package
	module, err = opt.Optimize(module)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error optimizing IR: %v\n", err)
		os.Exit(1)
	}

	// Generate machine code
	codegen := codegen.NewCodeGenerator(
		codegen.DefaultTarget(),
		"build",
	)

	if err := codegen.Generate(module); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}
}
