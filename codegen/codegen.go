// codegen/codegen.go
package codegen

import (
	"coolz-compiler/ast"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

type CodeGenerator struct {
	module *ir.Module

	// Runtime functions
	printf *ir.Func
	scanf  *ir.Func
	malloc *ir.Func
	free   *ir.Func
	gets   *ir.Func
	atoi   *ir.Func
}

func NewCodeGenerator() *CodeGenerator {
	cg := &CodeGenerator{
		module: ir.NewModule(),
	}

	// Declare external functions
	cg.printf = cg.declarePrintf()
	cg.scanf = cg.declareScanf()
	cg.malloc = cg.declareMalloc()
	cg.free = cg.declareFree()
	cg.gets = cg.declareGets()
	cg.atoi = cg.declareAtoi()

	return cg
}

func (cg *CodeGenerator) Generate(program *ast.Program) string {
	// Generate IO class methods
	cg.generateIOClass()

	// Generate Main class
	mainClass := cg.findMainClass(program)
	if mainClass != nil {
		cg.generateMainClass(mainClass)
	}

	return cg.module.String()
}

func (cg *CodeGenerator) generateIOClass() {
	// Generate IO class methods
	cg.generateOutString()
	cg.generateOutInt()
	cg.generateInString()
	cg.generateInInt()
}

func (cg *CodeGenerator) generateOutString() {
	fn := cg.module.NewFunc("IO_out_string",
		types.I8Ptr, // return type (SELF_TYPE as i8*)
		ir.NewParam("self", types.I8Ptr),
		ir.NewParam("str", types.I8Ptr),
	)

	block := fn.NewBlock("entry")

	// printf format string
	fmtStr := block.NewAlloca(types.I8)
	block.NewStore(constant.NewCharArrayFromString("%s\000"), fmtStr)

	// Call printf
	block.NewCall(cg.printf, fmtStr, fn.Params[1])

	// Return self
	block.NewRet(fn.Params[0])
}

func (cg *CodeGenerator) generateOutInt() {
	fn := cg.module.NewFunc("IO_out_int",
		types.I8Ptr,
		ir.NewParam("self", types.I8Ptr),
		ir.NewParam("num", types.I32),
	)

	block := fn.NewBlock("entry")

	fmtStr := block.NewAlloca(types.I8)
	block.NewStore(constant.NewCharArrayFromString("%d\000"), fmtStr)

	block.NewCall(cg.printf, fmtStr, fn.Params[1])
	block.NewRet(fn.Params[0])
}

func (cg *CodeGenerator) generateInString() {
	fn := cg.module.NewFunc("IO_in_string",
		types.I8Ptr,
		ir.NewParam("self", types.I8Ptr),
	)

	block := fn.NewBlock("entry")

	// Allocate buffer
	buf := block.NewCall(cg.malloc, constant.NewInt(types.I64, 256))
	bufPtr := block.NewBitCast(buf, types.I8Ptr)

	// Call gets
	block.NewCall(cg.gets, bufPtr)

	// Return buffer
	block.NewRet(bufPtr)
}

func (cg *CodeGenerator) generateInInt() {
	fn := cg.module.NewFunc("IO_in_int",
		types.I32,
		ir.NewParam("self", types.I8Ptr),
	)

	block := fn.NewBlock("entry")

	// Allocate buffer
	buf := block.NewCall(cg.malloc, constant.NewInt(types.I64, 256))
	bufPtr := block.NewBitCast(buf, types.I8Ptr)

	// Call gets
	block.NewCall(cg.gets, bufPtr)

	// Convert to int
	result := block.NewCall(cg.atoi, bufPtr)

	// Free buffer
	block.NewCall(cg.free, buf)

	block.NewRet(result)
}

// Helper functions to declare C stdlib functions
func (cg *CodeGenerator) declarePrintf() *ir.Func {
	printfType := types.NewFunc(types.I32, types.I8Ptr)
	printfType.Variadic = true
	return cg.module.NewFunc("printf", printfType)
}

func (cg *CodeGenerator) declareScanf() *ir.Func {
	scanfType := types.NewFunc(types.I32, types.I8Ptr)
	scanfType.Variadic = true
	return cg.module.NewFunc("scanf", scanfType)
}

func (cg *CodeGenerator) declareMalloc() *ir.Func {
	return cg.module.NewFunc("malloc",
		types.I8Ptr,
		ir.NewParam("size", types.I64),
	)
}

func (cg *CodeGenerator) declareFree() *ir.Func {
	return cg.module.NewFunc("free",
		types.Void,
		ir.NewParam("ptr", types.I8Ptr),
	)
}

func (cg *CodeGenerator) declareGets() *ir.Func {
	return cg.module.NewFunc("gets",
		types.I8Ptr,
		ir.NewParam("str", types.I8Ptr),
	)
}

func (cg *CodeGenerator) declareAtoi() *ir.Func {
	return cg.module.NewFunc("atoi",
		types.I32,
		ir.NewParam("str", types.I8Ptr),
	)
}

func (cg *CodeGenerator) findMainClass(program *ast.Program) *ast.Class {
	for _, class := range program.Classes {
		if class.Name.Value == "Main" {
			return class
		}
	}
	return nil
}

func (cg *CodeGenerator) generateMainClass(mainClass *ast.Class) {
	mainFn := cg.module.NewFunc("main", types.I32)
	block := mainFn.NewBlock("entry")

	// Locate "Main_main" function by name
	for _, fn := range cg.module.Funcs {
		if fn.Name() == "Main_main" {
			block.NewCall(fn)
			break
		}
	}

	block.NewRet(constant.NewInt(types.I32, 0))
}
