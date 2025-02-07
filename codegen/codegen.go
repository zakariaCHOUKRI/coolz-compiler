package codegen

import (
	"fmt"
	"strings"
)

type CodeGenerator struct {
	assembly     strings.Builder
	labelCounter int
	stackOffset  int
	stringIndex  int
	dataBuilder  strings.Builder
}

func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		labelCounter: 0,
		stackOffset:  0,
		stringIndex:  0,
	}
}

func (cg *CodeGenerator) Generate(ast interface{}) string {
	// Reset builders
	cg.assembly.Reset()
	cg.dataBuilder.Reset()

	// Generate code first (this will collect strings in dataBuilder)
	cg.generateCode(ast)

	// Now build the final assembly with proper sections
	finalAssembly := new(strings.Builder)

	// Data section
	finalAssembly.WriteString(".data\n")
	finalAssembly.WriteString(cg.dataBuilder.String())

	// Text section
	finalAssembly.WriteString(".text\n")
	finalAssembly.WriteString(".globl main\n")
	finalAssembly.WriteString(cg.assembly.String())

	return finalAssembly.String()
}

func (cg *CodeGenerator) generateCode(node interface{}) {
	switch n := node.(type) {
	case *Program:
		cg.generateProgram(n)
	case *Class:
		cg.generateClass(n)
	case *Method:
		cg.generateMethod(n)
	case *Attribute:
		cg.generateAttribute(n)
	// Concrete expression types first
	case *IntConstant:
		cg.generateIntConstant(n)
	case *StringConstant:
		cg.generateStringConstant(n)
	case *BoolConstant:
		cg.generateBoolConstant(n)
	case *Dispatch:
		cg.generateDispatch(n)
	case *If:
		cg.generateIf(n)
	case *While:
		cg.generateWhile(n)
	case *Block:
		cg.generateBlock(n)
	// General Expression interface last
	case Expression:
		cg.generateExpression(n)
	default:
		panic("Unknown node type in code generation")
	}
}

func (cg *CodeGenerator) generateProgram(program *Program) {
	for _, class := range program.Classes {
		cg.generateClass(class)
	}
}

func (cg *CodeGenerator) generateClass(class *Class) {
	// Generate vtable for the class
	cg.emit(".data")
	cg.emit("_vtable_" + class.Name + ":")

	// Generate code for each feature
	for _, feature := range class.Features {
		switch f := feature.(type) {
		case *Method:
			cg.generateMethod(f)
		case *Attribute:
			cg.generateAttribute(f)
		}
	}
}

func (cg *CodeGenerator) generateAttribute(attr *Attribute) {
	if attr.Init != nil {
		cg.generateExpression(attr.Init)
	} else {
		// Initialize with default value based on type
		switch attr.Type {
		case "Int":
			cg.emit("\tli $a0, 0")
		case "Bool":
			cg.emit("\tli $a0, 0")
		case "String":
			cg.emit("\tla $a0, empty_str")
		}
	}
}

func (cg *CodeGenerator) emit(code string) {
	cg.assembly.WriteString(code + "\n")
}

func (cg *CodeGenerator) newLabel() string {
	cg.labelCounter++
	return fmt.Sprintf("L%d", cg.labelCounter)
}
