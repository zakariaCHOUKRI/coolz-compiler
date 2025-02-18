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
	cg := codegen.NewCodeGenerator()
	irCode := cg.Generate(program)

	// Use outputName for both the executable and intermediate file
	llvmIR := filepath.Join(filepath.Dir(*outputName), "output.ll")
	if err := os.WriteFile(llvmIR, []byte(irCode), 0644); err != nil {
		panic(err)
	}

	// Compile with MSYS2 Clang using specified output name
	cmd := exec.Command("clang",
		llvmIR,
		"-o", *outputName,
		"-fuse-ld=lld",
		"-Wl,/subsystem:console",
		"-ladvapi32",
		"-lshell32",
		"-luser32",
		"-lkernel32",
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
