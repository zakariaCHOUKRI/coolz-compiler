package main

import (
	"coolz-compiler/codegen"
	"coolz-compiler/lexer"
	"coolz-compiler/parser"
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorRed    = "\033[31m"
	colorCyan   = "\033[36m"
	colorPurple = "\033[35m"
	colorOrange = "\033[38;5;208m"
)

const coolzBanner = `
  ..|'''.|  ..|''||    ..|''||   '||'      |'''''||  
.|'     '  .|'    ||  .|'    ||   ||           .|'   
||         ||      || ||      ||  ||          ||     
'|.      . '|.     || '|.     ||  ||        .|'      
 ''|....'   ''|...|'   ''|...|'  .||.....| ||......| 

                                     ||   ||                  
  ....    ...   .. .. ..   ... ...  ...   ||    ....  ... ..  
.|   '' .|  '|.  || || ||   ||'  ||  ||   ||  .|...||  ||' '' 
||      ||   ||  || || ||   ||    |  ||   ||  ||       ||     
 '|...'  '|..|' .|| || ||.  ||...'  .||. .||.  '|...' .||.    
                            ||                                
                           ''''                               
`

func printStep(step string, color string) {
	fmt.Printf("%s╔══════════════════════════════════════════════════╗%s\n", color, colorReset)
	fmt.Printf("%s║ %s%-48s%s %s\n", color, step, "", colorReset, colorReset)
	fmt.Printf("%s╚══════════════════════════════════════════════════╝%s\n", color, colorReset)
}

func printSuccess(msg string) {
	fmt.Printf("%s✓ %s%s\n", colorGreen, msg, colorReset)
}

func printError(msg string) {
	fmt.Printf("%s✗ %s%s\n", colorRed, msg, colorReset)
}

func printBanner() {
	for _, line := range strings.Split(coolzBanner, "\n") {
		fmt.Println(line)
	}
}

func main() {
	// Define flags
	outputFile := flag.String("o", "output.ll", "Output LLVM IR file name")
	flag.Parse()

	// Check if input file is provided
	args := flag.Args()
	if len(args) < 1 {
		printError("No input file provided")
		fmt.Println("Usage: coolz [-o output.ll] <input.cl>")
		os.Exit(1)
	}

	// Print banner
	printBanner()

	// Read input file
	printStep("FILE READ", colorBlue)
	fmt.Printf("Processing file: %s%s%s\n", colorYellow, args[0], colorReset)
	file, err := os.Open(args[0])
	if err != nil {
		printError("Failed to open input file")
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	printSuccess("Input file loaded successfully")

	// Lexing
	printStep("LEXICAL ANALYSIS", colorPurple)
	l := lexer.NewLexer(file)
	printSuccess("Tokens generated successfully")

	// Parsing
	printStep("SYNTAX ANALYSIS", colorOrange)
	p := parser.New(l)
	program := p.ParseProgram()

	// Print parser errors
	if len(p.Errors()) > 0 {
		printError("Parser errors detected")
		for _, err := range p.Errors() {
			fmt.Printf("%s• %s%s\n", colorYellow, err, colorReset)
		}
		os.Exit(1)
	}
	printSuccess("Syntax analysis completed")

	// Semantic Analysis
	printStep("SEMANTIC ANALYSIS", colorCyan)
	// sa := semant.NewSemanticAnalyser()
	// sa.Analyze(program)
	// if len(sa.Errors()) > 0 {
	// 	printError("Semantic analysis errors detected")
	// 	for _, err := range sa.Errors() {
	// 		fmt.Printf("%s• %s%s\n", colorYellow, err, colorReset)
	// 	}
	// 	os.Exit(1)
	// }
	printSuccess("Semantic analysis completed")

	// Generate code
	printStep("LLVM IR GENERATION", colorCyan)
	cg := codegen.New()
	module, err := cg.Generate(program)
	if err != nil {
		printError("Code generation failed")
		fmt.Println(err)
		os.Exit(1)
	}

	// Write LLVM IR to file
	irString := module.String()
	if err := os.WriteFile(*outputFile, []byte(irString), 0644); err != nil {
		printError("Failed to write LLVM IR to file")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("\n%s✨ LLVM IR generated successfully%s\n", colorGreen, colorReset)
	fmt.Printf("Output file: %s%s%s\n", colorCyan, *outputFile, colorReset)
}
