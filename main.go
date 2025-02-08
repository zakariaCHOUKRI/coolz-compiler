package main

import (
	"coolz-compiler/codegen"
	"coolz-compiler/irgen"
	"coolz-compiler/lexer"
	"coolz-compiler/optimizer"
	"coolz-compiler/parser"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <sourcefile.cl>\n", os.Args[0])
		os.Exit(1)
	}

	sourceFile := os.Args[1]
	contentBytes, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", sourceFile, err)
		os.Exit(1)
	}

	lex := lexer.NewLexer(strings.NewReader(string(contentBytes)))
	p := parser.New(lex)
	programAST := p.ParseProgram()
	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			fmt.Fprintln(os.Stderr, "Parse error:", e)
		}
		os.Exit(1)
	}

	generator := irgen.NewIRGenerator()
	module, err := generator.Generate(programAST)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating IR: %v\n", err)
		os.Exit(1)
	}

	opt := optimizer.NewOptimizer()
	opt.SetLevel(optimizer.MediumOptimization)
	module, err = opt.Optimize(module)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error optimizing IR: %v\n", err)
		os.Exit(1)
	}

	codegen := codegen.NewCodeGenerator(
		codegen.DefaultTarget(),
		"build",
	)
	if err := codegen.Generate(module); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}
}
