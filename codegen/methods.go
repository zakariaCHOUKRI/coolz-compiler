package codegen

import "fmt"

func (cg *CodeGenerator) generateMethod(method *Method) {
	// Method prologue
	cg.emit(fmt.Sprintf("%s:", method.Name))
	cg.emit("\tsw $fp, 0($sp)")
	cg.emit("\tsw $ra, -4($sp)")
	cg.emit("\tsw $s0, -8($sp)")
	cg.emit("\tmove $fp, $sp")
	cg.emit("\taddiu $sp, $sp, -12")

	// Generate method body
	cg.generateExpression(method.Body)

	// Method epilogue
	cg.emit("\tlw $fp, 0($sp)")
	cg.emit("\tlw $ra, -4($sp)")
	cg.emit("\tlw $s0, -8($sp)")
	cg.emit("\taddiu $sp, $sp, 12")
	cg.emit("\tjr $ra")
}

func (cg *CodeGenerator) generateDispatch(dispatch *Dispatch) {
	// Save caller-saved registers
	cg.emit("\tsw $a0, 0($sp)")
	cg.emit("\taddiu $sp, $sp, -4")

	// Evaluate and push arguments
	for _, arg := range dispatch.Arguments {
		cg.generateExpression(arg)
		cg.emit("\tsw $a0, 0($sp)")
		cg.emit("\taddiu $sp, $sp, -4")
	}

	// Call method
	cg.emit(fmt.Sprintf("\tjal %s", dispatch.MethodName))

	// Restore stack
	cg.emit(fmt.Sprintf("\taddiu $sp, $sp, %d", 4*len(dispatch.Arguments)))

	// Restore caller-saved registers
	cg.emit("\tlw $a0, 0($sp)")
	cg.emit("\taddiu $sp, $sp, 4")
}
