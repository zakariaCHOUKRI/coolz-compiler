package main

import (
	"coolz-compiler/codegen"
	"coolz-compiler/lexer"
	"coolz-compiler/parser"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Define flags
	outputName := flag.String("o", "a.exe", "Output executable name")
	flag.Parse()

	// Check if input file is provided
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: coolc [-o output] <input.cl>")
		os.Exit(1)
	}

	// Read input file
	file, err := os.Open(args[0])
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
	cg := codegen.New()
	module, err := cg.Generate(program)
	if err != nil {
		fmt.Println("Code generation error:", err)
		os.Exit(1)
	}

	// Write LLVM IR to file
	irString := module.String()
	llvmIR := filepath.Join(filepath.Dir(*outputName), "output.ll")
	if err := os.WriteFile(llvmIR, []byte(irString), 0644); err != nil {
		panic(err)
	}

	// Compile with MSYS2 Clang using specified output name
	cmd := exec.Command("clang-cl",
		llvmIR,
		//"-v",
		"/Fe:"+*outputName,
		"/link",
		"/subsystem:console",
		"advapi32.lib",
		"shell32.lib",
		"user32.lib",
		"kernel32.lib",
		"msvcrt.lib",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("Executing command:", cmd.String())
	if err := cmd.Run(); err != nil {
		fmt.Println("Linking failed:", err)
		os.Exit(1)
	}

	fmt.Printf("Compilation successful. Output: %s\n", *outputName)
}
