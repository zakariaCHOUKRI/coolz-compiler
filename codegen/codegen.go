package codegen

import (
	"coolz-compiler/ast"
	"fmt"
	"strings"

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
	scanf           *ir.Func
	methods         map[string]map[string]*ir.Func
	memset          *ir.Func
	currentBindings map[string]value.Value
}

// New creates a new code generator
func New() *CodeGenerator {
	cg := &CodeGenerator{
		module:          ir.NewModule(),
		stringConstants: make(map[string]*ir.Global),
		methods:         make(map[string]map[string]*ir.Func),
		currentBindings: make(map[string]value.Value),
	}

	// Set target triple for Windows MSVC
	cg.module.TargetTriple = "x86_64-pc-windows-msvc"

	// Declare external printf function
	printfType := types.NewPointer(types.I8)
	cg.printf = cg.module.NewFunc("printf", types.I32, ir.NewParam("format", printfType))
	cg.printf.Sig.Variadic = true

	// Declare external scanf function
	scanfType := types.NewPointer(types.I8)
	cg.scanf = cg.module.NewFunc("scanf", types.I32, ir.NewParam("format", scanfType))
	cg.scanf.Sig.Variadic = true

	// Declare external memset function
	memsetType := types.NewPointer(types.I8)
	cg.memset = cg.module.NewFunc("memset", types.NewPointer(types.I8),
		ir.NewParam("str", memsetType),
		ir.NewParam("c", types.I32),
		ir.NewParam("n", types.I64))

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
	block.NewRet(nil)

	// Create in_string method
	inString := cg.module.NewFunc("IO_in_string", types.NewPointer(types.I8),
		ir.NewParam("self", types.NewPointer(types.I8)))
	cg.methods["IO"]["in_string"] = inString

	block = inString.NewBlock("")

	// Allocate buffer for input string (256 bytes should be enough)
	buffer := block.NewAlloca(types.NewArray(256, types.I8))

	// Create format string for reading string
	strFormat = cg.getStringConstant("%255[^\n]")

	// Clear input buffer first (set to all zeros)
	zero := constant.NewInt(types.I32, 0)
	strPtr := block.NewGetElementPtr(types.NewArray(256, types.I8), buffer, zero, zero)
	block.NewCall(cg.memset, strPtr, zero, constant.NewInt(types.I64, 256))

	// Read the string
	block.NewCall(cg.scanf, strFormat, buffer)

	// Get the string length using strlen
	strlen := cg.module.NewFunc("strlen", types.I64,
		ir.NewParam("str", types.NewPointer(types.I8)))

	strPtr = block.NewGetElementPtr(types.NewArray(256, types.I8), buffer, zero, zero)
	length := block.NewCall(strlen, strPtr)

	// Allocate permanent storage for the string (+1 for null terminator)
	size := block.NewAdd(length, constant.NewInt(types.I64, 1))
	malloc := cg.module.NewFunc("malloc", types.NewPointer(types.I8),
		ir.NewParam("size", types.I64))
	permanent := block.NewCall(malloc, size)

	// Copy the string to permanent storage
	memcpy := cg.module.NewFunc("memcpy", types.NewPointer(types.I8),
		ir.NewParam("dest", types.NewPointer(types.I8)),
		ir.NewParam("src", types.NewPointer(types.I8)),
		ir.NewParam("size", types.I64))
	block.NewCall(memcpy, permanent, strPtr, size)

	// Return the permanent string
	block.NewRet(permanent)

	// Create in_int method
	inInt := cg.module.NewFunc("IO_in_int", types.I64,
		ir.NewParam("self", types.NewPointer(types.I8)))
	cg.methods["IO"]["in_int"] = inInt

	block = inInt.NewBlock("")

	// Allocate space for the integer
	intVar := block.NewAlloca(types.I64)

	// Create format string for reading integer
	intFormat = cg.getStringConstant("%lld")

	// Read integer
	block.NewCall(cg.scanf, intFormat, intVar)

	// Clear input buffer properly
	block.NewCall(cg.scanf, cg.getStringConstant("%*c"))

	// Load and return the integer
	result := block.NewLoad(types.I64, intVar)
	block.NewRet(result)

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
	case *ast.LetExpression:
		return cg.generateLet(block, e)
	case *ast.ObjectIdentifier:
		// Look up the variable in current bindings
		for name, alloca := range cg.currentBindings {
			if strings.HasPrefix(name, e.Value) {
				// Load the value from the alloca instruction
				return block.NewLoad(alloca.Type().(*types.PointerType).ElemType, alloca), nil
			}
		}
		return nil, fmt.Errorf("undefined variable: %s", e.Value)
	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func (cg *CodeGenerator) generateLet(block *ir.Block, letExpr *ast.LetExpression) (value.Value, error) {
	// Store the previous bindings map to restore later
	prevBindings := make(map[string]value.Value)
	for k, v := range cg.currentBindings {
		prevBindings[k] = v
	}

	// Process each binding
	for i, binding := range letExpr.Bindings {
		uniqueName := fmt.Sprintf("%s_let%d_%d", binding.Identifier.Value, len(cg.currentBindings), i)

		// Allocate space for the variable
		varType := cg.getLLVMType(binding.Type)
		alloca := block.NewAlloca(varType)

		if binding.Init != nil {
			initValue, err := cg.generateExpression(block, binding.Init)
			if err != nil {
				return nil, err
			}
			block.NewStore(initValue, alloca)
		} else {
			var defaultValue value.Value
			switch binding.Type.Value {
			case "Int":
				defaultValue = constant.NewInt(types.I64, 0)
			case "Bool":
				defaultValue = constant.NewBool(false)
			case "String":
				defaultValue = constant.NewNull(types.NewPointer(types.I8))
			default:
				defaultValue = constant.NewNull(types.NewPointer(types.I8))
			}
			block.NewStore(defaultValue, alloca)
		}

		cg.currentBindings[uniqueName] = alloca
	}

	// Generate code for the body expression
	result, err := cg.generateExpression(block, letExpr.In)

	// Restore previous bindings
	cg.currentBindings = prevBindings

	return result, err
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

		// Only add return instruction if this is the last expression
		// AND we're in a method body (not in a let expression or other block)
		if i == len(blockExpr.Expressions)-1 && block.Term == nil {
			block.NewRet(lastValue)
		}
	}

	return lastValue, nil
}

// generateDispatch generates code for method dispatch
func (cg *CodeGenerator) generateDispatch(block *ir.Block, dispatch *ast.DynamicDispatch) (value.Value, error) {
	// Handle all IO methods
	methodName := dispatch.Method.Value
	if methodName != "out_string" && methodName != "out_int" &&
		methodName != "in_string" && methodName != "in_int" {
		return nil, fmt.Errorf("unsupported method: %s", methodName)
	}

	// Generate code for arguments
	args := make([]value.Value, 0, len(dispatch.Arguments)+1)

	// Add self parameter
	args = append(args, constant.NewNull(types.NewPointer(types.I8)))

	// For output methods, process the argument
	if methodName == "out_string" || methodName == "out_int" {
		if len(dispatch.Arguments) != 1 {
			return nil, fmt.Errorf("expected 1 argument for %s, got %d", methodName, len(dispatch.Arguments))
		}

		arg, err := cg.generateExpression(block, dispatch.Arguments[0])
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	} else {
		// For input methods, verify no arguments
		if len(dispatch.Arguments) != 0 {
			return nil, fmt.Errorf("expected 0 arguments for %s, got %d", methodName, len(dispatch.Arguments))
		}
	}

	// Call the appropriate IO method
	method := cg.methods["IO"][methodName]
	result := block.NewCall(method, args...)

	// For input methods, return the result. For output methods, return null
	if methodName == "in_string" || methodName == "in_int" {
		return result, nil
	}
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
