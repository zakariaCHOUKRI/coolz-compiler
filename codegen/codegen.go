package codegen

import (
	"coolz-compiler/ast"
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

// CodeGenerator holds the state needed for code generation
type CodeGenerator struct {
	module          *ir.Module
	currentFunc     *ir.Func
	stringConstants map[string]*ir.Global
	printf          *ir.Func
	methods         map[string]map[string]*ir.Func
}

// New creates a new code generator
func New() *CodeGenerator {
	cg := &CodeGenerator{
		module:          ir.NewModule(),
		stringConstants: make(map[string]*ir.Global),
		methods:         make(map[string]map[string]*ir.Func),
	}

	// Set target triple for Windows MSVC
	cg.module.TargetTriple = "x86_64-pc-windows-msvc"

	// Declare external printf function
	printfType := types.NewPointer(types.I8)
	cg.printf = cg.module.NewFunc("printf", types.I32, ir.NewParam("format", printfType))
	cg.printf.Sig.Variadic = true

	return cg
}

// Generate generates LLVM IR for the entire program
func (cg *CodeGenerator) Generate(program *ast.Program) (*ir.Module, error) {
	// Initialize IO class methods
	cg.methods["IO"] = make(map[string]*ir.Func)

	// Create out_string method
	outString := cg.module.NewFunc("IO_out_string", types.Void,
		ir.NewParam("self", types.NewPointer(types.I8)),
		ir.NewParam("x", types.NewPointer(types.I8)))
	cg.methods["IO"]["out_string"] = outString

	block := outString.NewBlock("")

	// Create format string for printing strings
	strFormat := cg.getStringConstant("%s")

	// Call printf with the string
	block.NewCall(cg.printf, strFormat, outString.Params[1])
	// Print newline
	// newlineFormat := cg.getStringConstant("\n")
	// block.NewCall(cg.printf, newlineFormat)
	block.NewRet(nil)

	// Create out_int method
	outInt := cg.module.NewFunc("IO_out_int", types.Void,
		ir.NewParam("self", types.NewPointer(types.I8)),
		ir.NewParam("x", types.I64))
	cg.methods["IO"]["out_int"] = outInt

	block = outInt.NewBlock("")

	// Create format string for printing integers
	intFormat := cg.getStringConstant("%lld")

	// Call printf with the integer
	block.NewCall(cg.printf, intFormat, outInt.Params[1])
	// Print newline
	// block.NewCall(cg.printf, newlineFormat)
	block.NewRet(nil)

	// Generate code for all classes
	for _, class := range program.Classes {
		err := cg.generateClass(class)
		if err != nil {
			return nil, err
		}
	}

	// Generate main function
	mainFunc := cg.module.NewFunc("main", types.I32)
	block = mainFunc.NewBlock("")

	// Find Main class and main method
	var mainClass *ast.Class
	for _, class := range program.Classes {
		if class.Name.Value == "Main" {
			mainClass = class
			break
		}
	}

	if mainClass == nil {
		return nil, fmt.Errorf("no Main class found")
	}

	// Create new Main object (for now just using null as we haven't implemented object creation)
	mainObj := constant.NewNull(types.NewPointer(types.I8))

	// Call Main.main()
	mainMethod := cg.methods["Main"]["main"]
	if mainMethod == nil {
		return nil, fmt.Errorf("no main method found in Main class")
	}

	block.NewCall(mainMethod, mainObj)

	// Return 0 from main
	block.NewRet(constant.NewInt(types.I32, 0))

	return cg.module, nil
}

// getStringConstant creates or retrieves a global string constant
func (cg *CodeGenerator) getStringConstant(s string) value.Value {
	if global, exists := cg.stringConstants[s]; exists {
		return global
	}

	// Create new global string constant
	data := constant.NewCharArrayFromString(s + "\x00")
	global := cg.module.NewGlobalDef("str."+fmt.Sprintf("%d", len(cg.stringConstants)), data)
	cg.stringConstants[s] = global

	// Get pointer to the first character
	zero := constant.NewInt(types.I32, 0)
	return constant.NewGetElementPtr(global.ContentType, global, zero, zero)
}

// generateClass generates code for a single class
func (cg *CodeGenerator) generateClass(class *ast.Class) error {
	// Initialize method map for this class if it doesn't exist
	if _, exists := cg.methods[class.Name.Value]; !exists {
		cg.methods[class.Name.Value] = make(map[string]*ir.Func)
	}

	// Generate code for each feature (method or attribute)
	for _, feature := range class.Features {
		switch f := feature.(type) {
		case *ast.Method:
			err := cg.generateMethod(class.Name.Value, f)
			if err != nil {
				return err
			}
		case *ast.Attribute:
			// Attributes will be handled later when we implement object creation
			continue
		}
	}

	return nil
}

// generateExpression generates code for an expression
func (cg *CodeGenerator) generateExpression(block *ir.Block, expr ast.Expression) (value.Value, error) {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		return cg.getStringConstant(e.Value), nil
	case *ast.IntegerLiteral:
		return constant.NewInt(types.I64, e.Value), nil
	case *ast.DynamicDispatch:
		return cg.generateDispatch(block, e)
	case *ast.BlockExpression:
		return cg.generateBlock(block, e)
	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// generateBlock generates code for a block expression
func (cg *CodeGenerator) generateBlock(block *ir.Block, blockExpr *ast.BlockExpression) (value.Value, error) {
	var lastValue value.Value
	var err error

	// Generate code for each expression in the block
	for i, expr := range blockExpr.Expressions {
		lastValue, err = cg.generateExpression(block, expr)
		if err != nil {
			return nil, err
		}

		// If this is the last expression in the block
		if i == len(blockExpr.Expressions)-1 {
			// Add return instruction for the last value
			block.NewRet(lastValue)
		}
	}

	return lastValue, nil
}

// generateDispatch generates code for method dispatch
func (cg *CodeGenerator) generateDispatch(block *ir.Block, dispatch *ast.DynamicDispatch) (value.Value, error) {
	// For now, we only handle IO.out_string and IO.out_int
	methodName := dispatch.Method.Value
	if methodName != "out_string" && methodName != "out_int" {
		return nil, fmt.Errorf("unsupported method: %s", methodName)
	}

	// Generate code for arguments
	args := make([]value.Value, 0, len(dispatch.Arguments)+1)

	// Add self parameter
	args = append(args, constant.NewNull(types.NewPointer(types.I8)))

	// Generate code for the argument
	if len(dispatch.Arguments) != 1 {
		return nil, fmt.Errorf("expected 1 argument for %s, got %d", methodName, len(dispatch.Arguments))
	}

	arg, err := cg.generateExpression(block, dispatch.Arguments[0])
	if err != nil {
		return nil, err
	}
	args = append(args, arg)

	// Call the appropriate IO method
	method := cg.methods["IO"][methodName]
	block.NewCall(method, args...)

	// Return null as SELF_TYPE for IO methods
	return constant.NewNull(types.NewPointer(types.I8)), nil
}

// generateMethod generates code for a method
func (cg *CodeGenerator) generateMethod(className string, method *ast.Method) error {
	// Create function parameters
	params := make([]*ir.Param, 0, len(method.Formals)+1)

	// Add self parameter
	params = append(params, ir.NewParam("self", types.NewPointer(types.I8)))

	// Add formal parameters
	for _, formal := range method.Formals {
		paramType := cg.getLLVMType(formal.Type)
		params = append(params, ir.NewParam(formal.Name.Value, paramType))
	}

	// Create function
	returnType := cg.getLLVMType(method.Type)
	fn := cg.module.NewFunc(fmt.Sprintf("%s_%s", className, method.Name.Value),
		returnType, params...)

	cg.methods[className][method.Name.Value] = fn

	// Generate code for method body
	cg.currentFunc = fn
	block := fn.NewBlock("")

	// Generate expression
	value, err := cg.generateExpression(block, method.Body)
	if err != nil {
		return err
	}

	// Ensure the block is terminated if it doesn't have a terminator
	if block.Term == nil {
		block.NewRet(value)
	}

	return nil
}

// getLLVMType converts a COOL type to an LLVM type
func (cg *CodeGenerator) getLLVMType(typeId *ast.TypeIdentifier) types.Type {
	switch typeId.Value {
	case "Int":
		return types.I64
	case "String":
		return types.NewPointer(types.I8)
	case "Bool":
		return types.I1
	case "SELF_TYPE":
		return types.NewPointer(types.I8)
	default:
		// For now, treat all other types as opaque pointers
		return types.NewPointer(types.I8)
	}
}
