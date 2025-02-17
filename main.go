// main.go
package main

import (
	"coolz-compiler/ast"
	"coolz-compiler/codegen"
	"coolz-compiler/lexer"
	"coolz-compiler/parser"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: coolc <input.cl>")
		os.Exit(1)
	}

	// Read input file
	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Lexing
	l := lexer.NewLexer(file)

	// Parsing
	p := parser.New(l)
	program := p.ParseProgram()

	// Print parser errors
	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Println(" ", err)
		}
		os.Exit(1)
	}

	// Generate code
	cg := codegen.NewCodeGenerator()
	irCode := cg.Generate(program)

	// Write LLVM IR to file
	outputFile := "output.ll"
	if err := os.WriteFile(outputFile, []byte(irCode), 0644); err != nil {
		panic(err)
	}

	// Compile with clang
	cmd := exec.Command("clang", "-Wno-override-module", "-o", "a.out", outputFile)
	if err := cmd.Run(); err != nil {
		fmt.Println("Linking failed:", err)
		os.Exit(1)
	}

	fmt.Println("Compilation successful. Output: a.out")
}

// Dummy semantic analyzer to satisfy compiler steps
type dummySemanticAnalyzer struct{}

func (d *dummySemanticAnalyzer) Analyze(*ast.Program) {}
