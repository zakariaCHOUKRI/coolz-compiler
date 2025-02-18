package codegen

import (
	"coolz-compiler/ast"
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

// CodeGen represents the code generator for the Cool programming language.
type CodeGen struct {
	module *ir.Module
}

// NewCodeGen creates a new CodeGen instance.
func NewCodeGen() *CodeGen {
	return &CodeGen{
		module: ir.NewModule(),
	}
}

// GenerateIR generates LLVM IR for the given AST.
func (cg *CodeGen) GenerateIR(program *ast.Program) string {
	// Generate IR for each class in the program
	for _, class := range program.Classes {
		cg.generateClassIR(class)
	}

	// Return the generated IR as a string
	return cg.module.String()
}

// generateClassIR generates LLVM IR for a given class.
func (cg *CodeGen) generateClassIR(class *ast.Class) {
	// If the class is the IO class, generate the IO methods
	if class.Name.Value == "IO" {
		cg.generateIOIR()
	}

	// Generate IR for each feature in the class
	for _, feature := range class.Features {
		switch f := feature.(type) {
		case *ast.Method:
			cg.generateMethodIR(class, f)
		case *ast.Attribute:
			cg.generateAttributeIR(class, f)
		}
	}
}

// generateMethodIR generates LLVM IR for a given method.
func (cg *CodeGen) generateMethodIR(class *ast.Class, method *ast.Method) {
	// Create a new function for the method
	fn := cg.module.NewFunc(
		method.Name.Value,
		cg.getLLVMType(method.Type),
		cg.getMethodParams(method)...,
	)

	// Generate IR for the method body
	if method.Body != nil {
		entryBlock := fn.NewBlock("entry")
		cg.generateExpressionIR(entryBlock, method.Body)
		// Ensure the block ends with a terminator
		if retType, ok := fn.Sig.RetType.(*types.PointerType); ok {
			entryBlock.NewRet(constant.NewNull(retType))
		} else {
			// Handle non-pointer return types (e.g., integers, booleans)
			entryBlock.NewRet(constant.NewInt(types.I32, 0)) // Default to returning 0 for non-pointer types
		}
	}
}

// generateAttributeIR generates LLVM IR for a given attribute.
func (cg *CodeGen) generateAttributeIR(class *ast.Class, attr *ast.Attribute) {
	// TODO: Implement attribute code generation
}

// generateExpressionIR generates LLVM IR for a given expression.
func (cg *CodeGen) generateExpressionIR(block *ir.Block, expr ast.Expression) value.Value {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return constant.NewInt(types.I32, e.Value)
	case *ast.StringLiteral:
		return cg.generateStringLiteralIR(block, e)
	case *ast.BooleanLiteral:
		return constant.NewInt(types.I1, boolToInt(e.Value))
	case *ast.UnaryExpression:
		return cg.generateUnaryExpressionIR(block, e)
	case *ast.BinaryExpression:
		return cg.generateBinaryExpressionIR(block, e)
	case *ast.IfExpression:
		return cg.generateIfExpressionIR(block, e)
	case *ast.WhileExpression:
		return cg.generateWhileExpressionIR(block, e)
	case *ast.BlockExpression:
		return cg.generateBlockExpressionIR(block, e)
	case *ast.LetExpression:
		return cg.generateLetExpressionIR(block, e)
	case *ast.NewExpression:
		return cg.generateNewExpressionIR(block, e)
	case *ast.IsVoidExpression:
		return cg.generateIsVoidExpressionIR(block, e)
	case *ast.CaseExpression:
		return cg.generateCaseExpressionIR(block, e)
	case *ast.Assignment:
		return cg.generateAssignmentIR(block, e)
	case *ast.DynamicDispatch:
		return cg.generateDynamicDispatchIR(block, e)
	case *ast.StaticDispatch:
		return cg.generateStaticDispatchIR(block, e)
	case *ast.Self:
		return cg.generateSelfIR(block, e)
	case *ast.VoidLiteral:
		return cg.generateVoidLiteralIR(block, e)
	default:
		panic(fmt.Sprintf("Unsupported expression type: %T", e))
	}
}

// generateStringLiteralIR generates LLVM IR for a string literal.
func (cg *CodeGen) generateStringLiteralIR(block *ir.Block, str *ast.StringLiteral) value.Value {
	// TODO: Implement string literal code generation
	return nil
}

// generateUnaryExpressionIR generates LLVM IR for a unary expression.
func (cg *CodeGen) generateUnaryExpressionIR(block *ir.Block, expr *ast.UnaryExpression) value.Value {
	// TODO: Implement unary expression code generation
	return nil
}

// generateBinaryExpressionIR generates LLVM IR for a binary expression.
func (cg *CodeGen) generateBinaryExpressionIR(block *ir.Block, expr *ast.BinaryExpression) value.Value {
	// TODO: Implement binary expression code generation
	return nil
}

// generateIfExpressionIR generates LLVM IR for an if expression.
func (cg *CodeGen) generateIfExpressionIR(block *ir.Block, expr *ast.IfExpression) value.Value {
	// TODO: Implement if expression code generation
	return nil
}

// generateWhileExpressionIR generates LLVM IR for a while expression.
func (cg *CodeGen) generateWhileExpressionIR(block *ir.Block, expr *ast.WhileExpression) value.Value {
	// TODO: Implement while expression code generation
	return nil
}

// generateBlockExpressionIR generates LLVM IR for a block expression.
func (cg *CodeGen) generateBlockExpressionIR(block *ir.Block, expr *ast.BlockExpression) value.Value {
	// TODO: Implement block expression code generation
	return nil
}

// generateLetExpressionIR generates LLVM IR for a let expression.
func (cg *CodeGen) generateLetExpressionIR(block *ir.Block, expr *ast.LetExpression) value.Value {
	// TODO: Implement let expression code generation
	return nil
}

// generateNewExpressionIR generates LLVM IR for a new expression.
func (cg *CodeGen) generateNewExpressionIR(block *ir.Block, expr *ast.NewExpression) value.Value {
	// TODO: Implement new expression code generation
	return nil
}

// generateIsVoidExpressionIR generates LLVM IR for an isvoid expression.
func (cg *CodeGen) generateIsVoidExpressionIR(block *ir.Block, expr *ast.IsVoidExpression) value.Value {
	// TODO: Implement isvoid expression code generation
	return nil
}

// generateCaseExpressionIR generates LLVM IR for a case expression.
func (cg *CodeGen) generateCaseExpressionIR(block *ir.Block, expr *ast.CaseExpression) value.Value {
	// TODO: Implement case expression code generation
	return nil
}

// generateAssignmentIR generates LLVM IR for an assignment expression.
func (cg *CodeGen) generateAssignmentIR(block *ir.Block, expr *ast.Assignment) value.Value {
	// TODO: Implement assignment code generation
	return nil
}

// generateDynamicDispatchIR generates LLVM IR for a dynamic dispatch expression.
func (cg *CodeGen) generateDynamicDispatchIR(block *ir.Block, expr *ast.DynamicDispatch) value.Value {
	// TODO: Implement dynamic dispatch code generation
	return nil
}

// generateStaticDispatchIR generates LLVM IR for a static dispatch expression.
func (cg *CodeGen) generateStaticDispatchIR(block *ir.Block, expr *ast.StaticDispatch) value.Value {
	// TODO: Implement static dispatch code generation
	return nil
}

// generateSelfIR generates LLVM IR for a self expression.
func (cg *CodeGen) generateSelfIR(block *ir.Block, expr *ast.Self) value.Value {
	// TODO: Implement self expression code generation
	return nil
}

// generateVoidLiteralIR generates LLVM IR for a void literal.
func (cg *CodeGen) generateVoidLiteralIR(block *ir.Block, expr *ast.VoidLiteral) value.Value {
	// TODO: Implement void literal code generation
	return nil
}

// getLLVMType returns the LLVM type corresponding to the given Cool type.
func (cg *CodeGen) getLLVMType(typeID *ast.TypeIdentifier) types.Type {
	switch typeID.Value {
	case "Int":
		return types.I32
	case "Bool":
		return types.I1
	case "String":
		return types.NewPointer(types.I8)
	case "SELF_TYPE":
		return types.NewPointer(types.I8) // Assuming SELF_TYPE is a pointer to an object
	default:
		return types.NewPointer(types.I8) // Default to a pointer for objects
	}
}

// getMethodParams returns the LLVM function parameters for a given method.
func (cg *CodeGen) getMethodParams(method *ast.Method) []*ir.Param {
	params := make([]*ir.Param, 0, len(method.Formals)+1)
	// Add self parameter
	params = append(params, ir.NewParam("self", types.NewPointer(types.I8)))
	// Add formal parameters
	for _, formal := range method.Formals {
		params = append(params, ir.NewParam(formal.Name.Value, cg.getLLVMType(formal.Type)))
	}
	return params
}

// boolToInt converts a boolean value to an integer (1 for true, 0 for false).
func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// generateIOIR generates LLVM IR for the IO class methods.
func (cg *CodeGen) generateIOIR() {
	// Generate IR for out_string method
	cg.generateOutStringIR()

	// Generate IR for out_int method
	cg.generateOutIntIR()

	// Generate IR for in_string method
	cg.generateInStringIR()

	// Generate IR for in_int method
	cg.generateInIntIR()
}

// generateOutStringIR generates LLVM IR for the out_string method.
func (cg *CodeGen) generateOutStringIR() {
	// Create the out_string function
	fn := cg.module.NewFunc(
		"out_string",
		types.NewPointer(types.I8), // Return type is SELF_TYPE (assumed to be a pointer)
		ir.NewParam("self", types.NewPointer(types.I8)), // self parameter
		ir.NewParam("x", types.NewPointer(types.I8)),    // x parameter (String)
	)

	// Generate the function body
	entryBlock := fn.NewBlock("entry")
	// Call the printf function to print the string
	printf := cg.module.NewFunc(
		"printf",
		types.I32,
		ir.NewParam("format", types.NewPointer(types.I8)),
	)
	printf.Sig.Variadic = true
	formatStr := cg.module.NewGlobalDef("fmt_str", constant.NewCharArrayFromString("%s\n\x00"))
	entryBlock.NewCall(printf, formatStr, fn.Params[1])
	// Return self
	entryBlock.NewRet(fn.Params[0])
}

// generateOutIntIR generates LLVM IR for the out_int method.
func (cg *CodeGen) generateOutIntIR() {
	// Create the out_int function
	fn := cg.module.NewFunc(
		"out_int",
		types.NewPointer(types.I8), // Return type is SELF_TYPE (assumed to be a pointer)
		ir.NewParam("self", types.NewPointer(types.I8)), // self parameter
		ir.NewParam("x", types.I32),                     // x parameter (Int)
	)

	// Generate the function body
	entryBlock := fn.NewBlock("entry")
	// Call the printf function to print the integer
	printf := cg.module.NewFunc(
		"printf",
		types.I32,
		ir.NewParam("format", types.NewPointer(types.I8)),
	)
	printf.Sig.Variadic = true
	formatStr := cg.module.NewGlobalDef("fmt_int", constant.NewCharArrayFromString("%d\n\x00"))
	entryBlock.NewCall(printf, formatStr, fn.Params[1])
	// Return self
	entryBlock.NewRet(fn.Params[0])
}

// generateInStringIR generates LLVM IR for the in_string method.
func (cg *CodeGen) generateInStringIR() {
	// Create the in_string function
	fn := cg.module.NewFunc(
		"in_string",
		types.NewPointer(types.I8), // Return type is String (pointer to i8)
		ir.NewParam("self", types.NewPointer(types.I8)), // self parameter
	)

	// Generate the function body
	entryBlock := fn.NewBlock("entry")
	// Call the scanf function to read a string
	scanf := cg.module.NewFunc(
		"scanf",
		types.I32,
		ir.NewParam("format", types.NewPointer(types.I8)),
	)
	scanf.Sig.Variadic = true
	formatStr := cg.module.NewGlobalDef("fmt_str_in", constant.NewCharArrayFromString("%s\x00"))
	buffer := entryBlock.NewAlloca(types.NewArray(1024, types.I8))
	entryBlock.NewCall(scanf, formatStr, buffer)
	// Return the buffer
	entryBlock.NewRet(buffer)
}

// generateInIntIR generates LLVM IR for the in_int method.
func (cg *CodeGen) generateInIntIR() {
	// Create the in_int function
	fn := cg.module.NewFunc(
		"in_int",
		types.I32, // Return type is Int
		ir.NewParam("self", types.NewPointer(types.I8)), // self parameter
	)

	// Generate the function body
	entryBlock := fn.NewBlock("entry")
	// Call the scanf function to read an integer
	scanf := cg.module.NewFunc(
		"scanf",
		types.I32,
		ir.NewParam("format", types.NewPointer(types.I8)),
	)
	scanf.Sig.Variadic = true
	formatStr := cg.module.NewGlobalDef("fmt_int_in", constant.NewCharArrayFromString("%d\x00"))
	result := entryBlock.NewAlloca(types.I32)
	entryBlock.NewCall(scanf, formatStr, result)
	// Load and return the result
	loadedResult := entryBlock.NewLoad(types.I32, result)
	entryBlock.NewRet(loadedResult)
}
