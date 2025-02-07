package main

import (
	"coolz-compiler/ast"
	"coolz-compiler/irgen"
	"fmt"
	"os"
)

func main() {
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

	generator := irgen.NewIRGenerator()
	module, err := generator.Generate(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating IR: %v\n", err)
		os.Exit(1)
	}

	// Print the generated LLVM IR
	fmt.Println(module.String())

	// Optionally, write to a file
	file, err := os.Create("output.ll")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	file.WriteString(module.String())
}
