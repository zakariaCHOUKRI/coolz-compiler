package codegen

import (
	"fmt"
	"strings"
)

func (cg *CodeGenerator) generateExpression(expr interface{}) {
	switch e := expr.(type) {
	case *IntConstant:
		cg.generateIntConstant(e)
	case *StringConstant:
		cg.generateStringConstant(e)
	case *BoolConstant:
		cg.generateBoolConstant(e)
	case *Dispatch:
		cg.generateDispatch(e)
	case *If:
		cg.generateIf(e)
	case *While:
		cg.generateWhile(e)
	case *Block:
		cg.generateBlock(e)
	}
}

func (cg *CodeGenerator) generateIntConstant(i *IntConstant) {
	cg.emit(fmt.Sprintf("\tli $a0, %d", i.Value))
}

// escapeString replaces newline characters with "\n" in the literal
func escapeString(s string) string {
	return strings.ReplaceAll(s, "\n", "\\n")
}

func (cg *CodeGenerator) createStringLabel() string {
	cg.stringIndex++
	return fmt.Sprintf("str%d", cg.stringIndex)
}

func (cg *CodeGenerator) generateStringConstant(s *StringConstant) {
	label := cg.createStringLabel()
	escaped := escapeString(s.Value)

	// Add to data section
	cg.dataBuilder.WriteString(fmt.Sprintf("%s: .asciiz \"%s\"\n", label, escaped))

	// Generate code to load the string
	cg.emit(fmt.Sprintf("\tla $a0, %s", label))
}

func (cg *CodeGenerator) generateBoolConstant(b *BoolConstant) {
	if b.Value {
		cg.emit("\tli $a0, 1")
	} else {
		cg.emit("\tli $a0, 0")
	}
}

func (cg *CodeGenerator) generateIf(ifExpr *If) {
	elseLabel := cg.newLabel()
	endLabel := cg.newLabel()

	// Generate condition code
	cg.generateExpression(ifExpr.Condition)
	cg.emit("\tbeqz $a0, " + elseLabel)

	// Generate then branch
	cg.generateExpression(ifExpr.ThenBranch)
	cg.emit("\tj " + endLabel)

	// Generate else branch
	cg.emit(elseLabel + ":")
	cg.generateExpression(ifExpr.ElseBranch)

	cg.emit(endLabel + ":")
}

func (cg *CodeGenerator) generateWhile(whileExpr *While) {
	startLabel := cg.newLabel()
	endLabel := cg.newLabel()

	cg.emit(startLabel + ":")
	// Generate condition code
	cg.generateExpression(whileExpr.Condition)
	cg.emit("\tbeqz $a0, " + endLabel)

	// Generate body
	cg.generateExpression(whileExpr.Body)
	cg.emit("\tj " + startLabel)

	cg.emit(endLabel + ":")
}

func (cg *CodeGenerator) generateBlock(block *Block) {
	for _, expr := range block.Expressions {
		cg.generateExpression(expr)
	}
}
