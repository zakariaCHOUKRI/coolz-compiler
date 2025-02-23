package codegen

import (
	"coolz-compiler/ast"
	"fmt"
	"sort"
	"strings"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
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
	blockCounter    int
	classes         map[string]*classInfo
	vtables         map[string]*ir.Global
	classTypes      map[string]*types.StructType
}

type classInfo struct {
	attributes     map[string]int    // Maps attribute names to their indices
	methods        map[string]int    // Maps method names to vtable indices
	attributeTypes []types.Type      // Types of attributes in order
	parent         string            // Name of parent class
	vtableType     *types.StructType // Type of the vtable
}

// New creates a new code generator
func New() *CodeGenerator {
	cg := &CodeGenerator{
		module:          ir.NewModule(),
		stringConstants: make(map[string]*ir.Global),
		methods:         make(map[string]map[string]*ir.Func),
		currentBindings: make(map[string]value.Value),
		classes:         make(map[string]*classInfo),
		vtables:         make(map[string]*ir.Global),
		classTypes:      make(map[string]*types.StructType),
	}

	// Set target triple for Windows MSVC
	cg.module.TargetTriple = "x86_64-pc-windows-msvc19.43.34808"

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

	// Register IO class first
	ioClass := &classInfo{
		attributes: make(map[string]int),
		methods:    make(map[string]int),
		parent:     "Object",
	}
	cg.classes["IO"] = ioClass

	// Register Object class (base class for all)
	objClass := &classInfo{
		attributes: make(map[string]int),
		methods:    make(map[string]int),
		parent:     "",
	}
	cg.classes["Object"] = objClass

	// First pass: Register all classes
	for _, class := range program.Classes {
		err := cg.registerClass(class)
		if err != nil {
			return nil, err
		}
	}

	// Second pass: Create vtables
	for _, class := range program.Classes {
		err := cg.createVTable(class.Name.Value)
		if err != nil {
			return nil, err
		}
	}

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

	// Third pass: Generate code for all classes
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

func (cg *CodeGenerator) generateClass(class *ast.Class) error {
	// Initialize method map for this class if it doesn't exist
	if _, exists := cg.methods[class.Name.Value]; !exists {
		cg.methods[class.Name.Value] = make(map[string]*ir.Func)
	}

	// First pass: Register all methods
	for _, feature := range class.Features {
		if method, ok := feature.(*ast.Method); ok {
			// Create function parameters
			params := make([]*ir.Param, 0, len(method.Formals)+1)

			// Add self parameter
			selfParam := ir.NewParam("self", types.NewPointer(types.I8))
			params = append(params, selfParam)

			// Add formal parameters
			for _, formal := range method.Formals {
				paramType := cg.getLLVMType(formal.Type)
				param := ir.NewParam(formal.Name.Value, paramType)
				params = append(params, param)
			}

			// Create function
			returnType := cg.getLLVMType(method.Type)
			fn := cg.module.NewFunc(fmt.Sprintf("%s_%s", class.Name.Value, method.Name.Value),
				returnType, params...)

			// Store the method in our method map
			cg.methods[class.Name.Value][method.Name.Value] = fn
		}
	}

	// Second pass: Generate method bodies
	for _, feature := range class.Features {
		switch f := feature.(type) {
		case *ast.Method:
			err := cg.generateMethodBody(class.Name.Value, f)
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

// Add this new function to handle method body generation
func (cg *CodeGenerator) generateMethodBody(className string, method *ast.Method) error {
	// Save the previous bindings
	prevBindings := cg.currentBindings
	cg.currentBindings = make(map[string]value.Value)

	// Get the previously created function
	fn := cg.methods[className][method.Name.Value]
	cg.currentFunc = fn

	// Create entry block
	block := fn.NewBlock("")

	// Add parameters to current bindings
	for i, formal := range method.Formals {
		// Create an alloca for the parameter
		alloca := block.NewAlloca(fn.Params[i+1].Type())
		// Store the parameter value
		block.NewStore(fn.Params[i+1], alloca)
		// Add to bindings
		cg.currentBindings[formal.Name.Value] = alloca
	}

	// Generate expression
	value, block, err := cg.generateExpression(block, method.Body)
	if err != nil {
		return err
	}

	// Ensure the block is terminated
	if block.Term == nil {
		block.NewRet(value)
	}

	// Restore previous bindings
	cg.currentBindings = prevBindings

	return nil
}

// generateExpression now returns (value, currentBlock, error)
// so that expressions which change control flow (like if) can update the current block.
func (cg *CodeGenerator) generateExpression(block *ir.Block, expr ast.Expression) (value.Value, *ir.Block, error) {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		return cg.getStringConstant(e.Value), block, nil
	case *ast.IntegerLiteral:
		return constant.NewInt(types.I64, e.Value), block, nil
	// In generateExpression, replace the DynamicDispatch case with:
	case *ast.DynamicDispatch:
		methodName := e.Method.Value

		// First check if this is an IO method
		if methodName == "out_string" || methodName == "out_int" ||
			methodName == "in_string" || methodName == "in_int" {

			// Existing IO method handling
			args := make([]value.Value, 0, len(e.Arguments)+1)
			args = append(args, constant.NewNull(types.NewPointer(types.I8))) // self

			if methodName == "out_string" || methodName == "out_int" {
				if len(e.Arguments) != 1 {
					return nil, block, fmt.Errorf("expected 1 argument for %s, got %d",
						methodName, len(e.Arguments))
				}
				arg, newBlock, err := cg.generateExpression(block, e.Arguments[0])
				if err != nil {
					return nil, block, err
				}
				block = newBlock
				args = append(args, arg)
			} else {
				if len(e.Arguments) != 0 {
					return nil, block, fmt.Errorf("expected 0 arguments for %s, got %d", methodName, len(e.Arguments))
				}
			}

			method := cg.methods["IO"][methodName]
			result := block.NewCall(method, args...)

			if methodName == "in_string" || methodName == "in_int" {
				return result, block, nil
			}
			return constant.NewNull(types.NewPointer(types.I8)), block, nil
		}

		// Handle regular method dispatch
		var obj value.Value
		var objClassName string

		if e.Object == nil || e.Object.TokenLiteral() == "self" {
			obj = cg.currentBindings["self"]
			objClassName = "Main" // Or the current class name
		} else {
			var err error
			var newBlock *ir.Block
			obj, newBlock, err = cg.generateExpression(block, e.Object)
			if err != nil {
				return nil, block, err
			}
			block = newBlock

			// Get the static type of the object from bindings
			if id, ok := e.Object.(*ast.ObjectIdentifier); ok {
				for name, binding := range cg.currentBindings {
					// Remove any unique suffixes from binding name
					baseName := strings.Split(name, "_let")[0]
					if baseName == id.Value {
						// Extract class name from binding type
						if ptrType, ok := binding.Type().(*types.PointerType); ok {
							if structType, ok := ptrType.ElemType.(*types.StructType); ok {
								objClassName = strings.Split(structType.Name(), ".")[0]
								break
							}
						}
					}
				}
				if objClassName == "" {
					return nil, block, fmt.Errorf("could not determine type for object %s", id.Value)
				}
			}
		}

		// Find and call the method through the class hierarchy
		method := cg.findMethodInHierarchy(objClassName, methodName)
		if method == nil {
			return nil, block, fmt.Errorf("method %s not found in class %s", methodName, objClassName)
		}

		// Generate arguments
		args := make([]value.Value, 0, len(e.Arguments)+1)
		args = append(args, obj) // self

		for _, arg := range e.Arguments {
			argValue, newBlock, err := cg.generateExpression(block, arg)
			if err != nil {
				return nil, block, err
			}
			block = newBlock
			args = append(args, argValue)
		}

		result := block.NewCall(method, args...)
		return result, block, nil
	case *ast.BlockExpression:
		return cg.generateBlock(block, e)
	case *ast.LetExpression:
		return cg.generateLet(block, e)
	case *ast.ObjectIdentifier:
		// Look up the variable in current bindings.
		for name, alloca := range cg.currentBindings {
			if strings.HasPrefix(name, e.Value) {
				// Load the value from the alloca instruction.
				return block.NewLoad(alloca.Type().(*types.PointerType).ElemType, alloca), block, nil
			}
		}
		return nil, block, fmt.Errorf("undefined variable: %s", e.Value)
	case *ast.BinaryExpression:
		left, block, err := cg.generateExpression(block, e.Left)
		if err != nil {
			return nil, block, err
		}
		right, block, err := cg.generateExpression(block, e.Right)
		if err != nil {
			return nil, block, err
		}
		switch e.Operator {
		// Arithmetic operations (both operands must be Int)
		case "+":
			return block.NewAdd(left, right), block, nil
		case "-":
			return block.NewSub(left, right), block, nil
		case "*":
			return block.NewMul(left, right), block, nil
		case "/":
			return block.NewSDiv(left, right), block, nil // Integer division only

		// Comparison operations
		case "<":
			return block.NewICmp(enum.IPredSLT, left, right), block, nil // Signed less than
		case "<=":
			return block.NewICmp(enum.IPredSLE, left, right), block, nil // Signed less than or equal
		case "=":
			return block.NewICmp(enum.IPredEQ, left, right), block, nil // Equal
		default:
			return nil, block, fmt.Errorf("unsupported binary operator: %s", e.Operator)
		}
	case *ast.UnaryExpression:
		operand, block, err := cg.generateExpression(block, e.Right)
		if err != nil {
			return nil, block, err
		}
		switch e.Operator {
		case "~": // Integer complement
			return block.NewSub(constant.NewInt(types.I64, 0), operand), block, nil
		case "not": // Boolean complement
			return block.NewXor(operand, constant.NewInt(types.I1, 1)), block, nil
		default:
			return nil, block, fmt.Errorf("unsupported unary operator: %s", e.Operator)
		}
	case *ast.IfExpression:
		// Generate code for the condition.
		condValue, block, err := cg.generateExpression(block, e.Condition)
		if err != nil {
			return nil, block, err
		}

		// Increment counter for unique block names.
		cg.blockCounter++
		thenBlock := cg.currentFunc.NewBlock(fmt.Sprintf("if_then_%d", cg.blockCounter))
		elseBlock := cg.currentFunc.NewBlock(fmt.Sprintf("if_else_%d", cg.blockCounter))
		mergeBlock := cg.currentFunc.NewBlock(fmt.Sprintf("if_merge_%d", cg.blockCounter))

		// Convert condition to i1 (boolean) if necessary.
		var condBool value.Value
		if !types.Equal(condValue.Type(), types.I1) {
			// For COOL, we consider non-zero values as true.
			condBool = block.NewICmp(enum.IPredNE, condValue, constant.NewInt(types.I64, 0))
		} else {
			condBool = condValue
		}

		// Create conditional branch.
		block.NewCondBr(condBool, thenBlock, elseBlock)

		// Generate code for then branch.
		thenValue, thenBlock, err := cg.generateExpression(thenBlock, e.Consequence)
		if err != nil {
			return nil, block, err
		}
		thenBlock.NewBr(mergeBlock)

		// Generate code for else branch.
		elseValue, elseBlock, err := cg.generateExpression(elseBlock, e.Alternative)
		if err != nil {
			return nil, block, err
		}
		elseBlock.NewBr(mergeBlock)

		// Check that types are compatible according to COOL's type system.
		if !types.Equal(thenValue.Type(), elseValue.Type()) {
			return nil, block, fmt.Errorf("type mismatch in if expression: then=%v else=%v",
				thenValue.Type(), elseValue.Type())
		}

		// Create PHI node in merge block.
		inc1 := &ir.Incoming{
			X:    thenValue,
			Pred: thenBlock,
		}
		inc2 := &ir.Incoming{
			X:    elseValue,
			Pred: elseBlock,
		}
		phi := mergeBlock.NewPhi(inc1)
		phi.Incs = append(phi.Incs, inc2)

		// Return the PHI node and update the current block to mergeBlock.
		return phi, mergeBlock, nil
	case *ast.WhileExpression:
		// Create the three blocks we need
		cg.blockCounter++
		condBlock := cg.currentFunc.NewBlock(fmt.Sprintf("while_cond_%d", cg.blockCounter))
		bodyBlock := cg.currentFunc.NewBlock(fmt.Sprintf("while_body_%d", cg.blockCounter))
		exitBlock := cg.currentFunc.NewBlock(fmt.Sprintf("while_exit_%d", cg.blockCounter))

		// Branch to condition block from current block
		block.NewBr(condBlock)

		// Generate condition code
		condValue, condBlock, err := cg.generateExpression(condBlock, e.Condition) // Fixed: Predicate -> Condition
		if err != nil {
			return nil, block, err
		}

		// Convert condition to boolean if necessary
		var condBool value.Value
		if !types.Equal(condValue.Type(), types.I1) {
			condBool = condBlock.NewICmp(enum.IPredNE, condValue, constant.NewInt(types.I64, 0))
		} else {
			condBool = condValue
		}

		// Create conditional branch
		condBlock.NewCondBr(condBool, bodyBlock, exitBlock)

		// Generate body code
		_, bodyBlock, err = cg.generateExpression(bodyBlock, e.Body)
		if err != nil {
			return nil, block, err
		}

		// Branch back to condition block from body
		bodyBlock.NewBr(condBlock)

		// Return void/null as the result of the loop
		return constant.NewNull(types.NewPointer(types.I8)), exitBlock, nil
	case *ast.Assignment:
		// Look up the variable in current bindings
		var alloca value.Value
		found := false
		for name, a := range cg.currentBindings {
			if strings.HasPrefix(name, e.Left.(*ast.ObjectIdentifier).Value) {
				alloca = a
				found = true
				break
			}
		}
		if !found {
			return nil, block, fmt.Errorf("undefined variable: %s", e.Left.(*ast.ObjectIdentifier).Value)
		}

		// Generate code for the value expression
		value, newBlock, err := cg.generateExpression(block, e.Value)
		if err != nil {
			return nil, block, err
		}

		// Store the new value
		newBlock.NewStore(value, alloca)

		// Return the value that was assigned
		return value, newBlock, nil
	default:
		return nil, block, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// generateBlock now threads the current block through each expression.
func (cg *CodeGenerator) generateBlock(block *ir.Block, blockExpr *ast.BlockExpression) (value.Value, *ir.Block, error) {
	var lastValue value.Value
	var err error
	currentBlock := block

	for _, expr := range blockExpr.Expressions {
		lastValue, currentBlock, err = cg.generateExpression(currentBlock, expr)
		if err != nil {
			return nil, currentBlock, err
		}
	}

	return lastValue, currentBlock, nil
}

func (cg *CodeGenerator) generateLet(block *ir.Block, letExpr *ast.LetExpression) (value.Value, *ir.Block, error) {
	// Store the previous bindings map to restore later
	prevBindings := make(map[string]value.Value)
	for k, v := range cg.currentBindings {
		prevBindings[k] = v
	}

	currentBlock := block

	// Process each binding
	for i, binding := range letExpr.Bindings {
		uniqueName := fmt.Sprintf("%s_let%d_%d", binding.Identifier.Value, len(cg.currentBindings), i)

		if binding.Type.Value != "Int" && binding.Type.Value != "Bool" && binding.Type.Value != "String" {
			if binding.Init != nil {
				initValue, newBlock, err := cg.generateExpression(currentBlock, binding.Init)
				if err != nil {
					return nil, currentBlock, err
				}
				currentBlock = newBlock
				cg.currentBindings[uniqueName] = initValue
			} else {
				// Create a new object of the specified type
				objValue, newBlock, err := cg.generateNew(currentBlock, binding.Type.Value)
				if err != nil {
					return nil, currentBlock, err
				}
				currentBlock = newBlock
				cg.currentBindings[uniqueName] = objValue
			}
		} else {

			// Allocate space for the variable
			varType := cg.getLLVMType(binding.Type)
			alloca := currentBlock.NewAlloca(varType)

			if binding.Init != nil {
				initValue, newBlock, err := cg.generateExpression(currentBlock, binding.Init)
				if err != nil {
					return nil, currentBlock, err
				}
				currentBlock = newBlock
				currentBlock.NewStore(initValue, alloca)
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
				currentBlock.NewStore(defaultValue, alloca)
			}

			cg.currentBindings[uniqueName] = alloca
		}
	}

	// Generate code for the body expression
	result, newBlock, err := cg.generateExpression(currentBlock, letExpr.In)

	// Restore previous bindings
	cg.currentBindings = prevBindings

	return result, newBlock, err
}

func (cg *CodeGenerator) generateDynamicDispatch(block *ir.Block, object value.Value, className string, methodName string, args []ast.Expression) (value.Value, *ir.Block, error) {
	// Find method in class hierarchy
	method := cg.findMethodInHierarchy(className, methodName)
	if method == nil {
		return nil, block, fmt.Errorf("method %s not found in class %s", methodName, className)
	}

	// Generate code for each argument
	llvmArgs := []value.Value{object} // First arg is always 'self'
	currentBlock := block

	for _, arg := range args {
		argValue, newBlock, err := cg.generateExpression(currentBlock, arg)
		if err != nil {
			return nil, currentBlock, err
		}
		currentBlock = newBlock
		llvmArgs = append(llvmArgs, argValue)
	}

	// Call the method
	result := currentBlock.NewCall(method, llvmArgs...)
	return result, currentBlock, nil
}

func (cg *CodeGenerator) findMethodInHierarchy(className string, methodName string) *ir.Func {
	current := className
	for current != "" {
		// Check if method exists in current class
		if methods, exists := cg.methods[current]; exists {
			if method, exists := methods[methodName]; exists {
				return method
			}
		}

		// Move up to parent class
		if classInfo, exists := cg.classes[current]; exists {
			current = classInfo.parent
		} else {
			current = ""
		}
	}
	return nil
}

func (cg *CodeGenerator) generateDispatch(block *ir.Block, dispatch *ast.DynamicDispatch) (value.Value, *ir.Block, error) {
	currentBlock := block

	// Handle all IO methods
	methodName := dispatch.Method.Value
	if methodName != "out_string" && methodName != "out_int" &&
		methodName != "in_string" && methodName != "in_int" {
		return nil, currentBlock, fmt.Errorf("unsupported method: %s", methodName)
	}

	// Generate code for arguments
	args := make([]value.Value, 0, len(dispatch.Arguments)+1)

	// Add self parameter
	args = append(args, constant.NewNull(types.NewPointer(types.I8)))

	// For output methods, process the argument
	if methodName == "out_string" || methodName == "out_int" {
		if len(dispatch.Arguments) != 1 {
			return nil, currentBlock, fmt.Errorf("expected 1 argument for %s, got %d", methodName, len(dispatch.Arguments))
		}

		arg, newBlock, err := cg.generateExpression(currentBlock, dispatch.Arguments[0])
		if err != nil {
			return nil, currentBlock, err
		}
		currentBlock = newBlock
		args = append(args, arg)
	} else {
		// For input methods, verify no arguments
		if len(dispatch.Arguments) != 0 {
			return nil, currentBlock, fmt.Errorf("expected 0 arguments for %s, got %d", methodName, len(dispatch.Arguments))
		}
	}

	// Call the appropriate IO method
	method := cg.methods["IO"][methodName]
	result := currentBlock.NewCall(method, args...)

	// For input methods, return the result. For output methods, return null
	if methodName == "in_string" || methodName == "in_int" {
		return result, currentBlock, nil
	}
	return constant.NewNull(types.NewPointer(types.I8)), currentBlock, nil
}

func (cg *CodeGenerator) generateMethod(className string, method *ast.Method) error {
	// Save the previous bindings
	prevBindings := cg.currentBindings
	cg.currentBindings = make(map[string]value.Value)

	// Create function parameters
	params := make([]*ir.Param, 0, len(method.Formals)+1)

	// Add self parameter
	selfParam := ir.NewParam("self", types.NewPointer(types.I8))
	params = append(params, selfParam)

	// Add formal parameters
	for _, formal := range method.Formals {
		paramType := cg.getLLVMType(formal.Type)
		param := ir.NewParam(formal.Name.Value, paramType)
		params = append(params, param)
	}

	// Create function
	returnType := cg.getLLVMType(method.Type)
	fn := cg.module.NewFunc(fmt.Sprintf("%s_%s", className, method.Name.Value),
		returnType, params...)

	// Store the method in our method map
	if cg.methods[className] == nil {
		cg.methods[className] = make(map[string]*ir.Func)
	}
	cg.methods[className][method.Name.Value] = fn

	// Generate code for method body
	cg.currentFunc = fn
	block := fn.NewBlock("")

	// Add parameters to current bindings
	for i, formal := range method.Formals {
		// Create an alloca for the parameter
		alloca := block.NewAlloca(params[i+1].Type())
		// Store the parameter value
		block.NewStore(params[i+1], alloca)
		// Add to bindings
		cg.currentBindings[formal.Name.Value] = alloca
	}

	// Generate expression
	value, block, err := cg.generateExpression(block, method.Body)
	if err != nil {
		return err
	}

	// Ensure the block is terminated
	if block.Term == nil {
		block.NewRet(value)
	}

	// Restore previous bindings
	cg.currentBindings = prevBindings

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

// Add to codegen.go
func (cg *CodeGenerator) generateMethodCall(block *ir.Block, object value.Value, className string,
	methodName string, args []ast.Expression) (value.Value, *ir.Block, error) {

	// Get the method from our method map
	method, exists := cg.methods[className][methodName]
	if !exists {
		return nil, block, fmt.Errorf("method %s not found in class %s", methodName, className)
	}

	// Generate code for each argument
	llvmArgs := []value.Value{object} // First arg is always 'self'
	currentBlock := block

	for _, arg := range args {
		argValue, newBlock, err := cg.generateExpression(currentBlock, arg)
		if err != nil {
			return nil, currentBlock, err
		}
		currentBlock = newBlock
		llvmArgs = append(llvmArgs, argValue)
	}

	// Call the method
	result := currentBlock.NewCall(method, llvmArgs...)
	return result, currentBlock, nil
}

// Add new method to register a class
func (cg *CodeGenerator) registerClass(class *ast.Class) error {
	className := class.Name.Value
	parentName := "Object"
	if class.Parent != nil {
		parentName = class.Parent.Value
	}

	// Check if class is already registered
	if _, exists := cg.classes[className]; exists {
		return nil // Already registered
	}

	// Check if parent exists and register it first if needed
	if _, exists := cg.classes[parentName]; !exists && parentName != "Object" {
		return fmt.Errorf("parent class %s not found for class %s", parentName, className)
	}

	info := &classInfo{
		attributes:     make(map[string]int),
		methods:        make(map[string]int),
		attributeTypes: make([]types.Type, 0),
		parent:         parentName,
	}

	// Inherit parent's attributes
	if parentInfo, exists := cg.classes[parentName]; exists {
		// Copy parent's attributes
		for attrName, idx := range parentInfo.attributes {
			info.attributes[attrName] = idx
			if idx < len(parentInfo.attributeTypes) {
				info.attributeTypes = append(info.attributeTypes, parentInfo.attributeTypes[idx])
			}
		}
	}

	// Add class's own attributes
	attrIndex := len(info.attributeTypes)
	for _, feature := range class.Features {
		if attr, ok := feature.(*ast.Attribute); ok {
			if _, exists := info.attributes[attr.Name.Value]; exists {
				return fmt.Errorf("attribute %s redefined in class %s", attr.Name.Value, className)
			}
			info.attributes[attr.Name.Value] = attrIndex
			info.attributeTypes = append(info.attributeTypes, cg.getLLVMType(attr.Type))
			attrIndex++
		}
	}

	// Initialize method map for this class
	cg.methods[className] = make(map[string]*ir.Func)

	// Create the class type including vtable pointer
	classType := types.NewStruct(append(
		[]types.Type{types.NewPointer(types.I8)}, // vtable pointer
		info.attributeTypes...,
	)...)
	cg.classTypes[className] = classType

	// Store the class info
	cg.classes[className] = info

	return nil
}

// Add new method to create vtable
func (cg *CodeGenerator) createVTable(className string) error {
	info := cg.classes[className]
	if info == nil {
		return fmt.Errorf("class %s not found", className)
	}

	// Collect all methods including inherited ones
	methods := make(map[string]*ir.Func)

	// Start from the current class and work up the inheritance chain
	current := className
	for current != "" {
		if classInfo, exists := cg.classes[current]; exists {
			// Add methods from current class, don't override existing ones
			if methodMap, exists := cg.methods[current]; exists {
				for name, method := range methodMap {
					if _, exists := methods[name]; !exists {
						methods[name] = method
					}
				}
			}
			current = classInfo.parent
		} else {
			break
		}
	}

	// Create vtable type and global
	methodTypes := make([]types.Type, 0, len(methods))
	methodValues := make([]constant.Constant, 0, len(methods))

	// Sort method names for consistent vtable layout
	methodNames := make([]string, 0, len(methods))
	for name := range methods {
		methodNames = append(methodNames, name)
	}
	sort.Strings(methodNames)

	for _, name := range methodNames {
		method := methods[name]
		info.methods[name] = len(methodTypes)
		methodTypes = append(methodTypes, types.NewPointer(method.Type()))
		methodValues = append(methodValues, method)
	}

	vtableType := types.NewStruct(methodTypes...)
	info.vtableType = vtableType

	vtable := cg.module.NewGlobalDef(
		fmt.Sprintf("%s_vtable", className),
		constant.NewStruct(vtableType, methodValues...),
	)
	cg.vtables[className] = vtable

	return nil
}

// Modified generateNew function with fixes for the type errors
func (cg *CodeGenerator) generateNew(block *ir.Block, className string) (value.Value, *ir.Block, error) {
	// Get class info and type
	info := cg.classes[className]
	if info == nil {
		return nil, block, fmt.Errorf("class %s not found", className)
	}
	classType := cg.classTypes[className]

	// Calculate size manually by getting size of each field
	var totalSize int64 = 8 // Start with 8 bytes for vtable pointer
	for _, attrType := range info.attributeTypes {
		switch attrType.(type) {
		case *types.IntType:
			totalSize += 8 // Int64
		case *types.FloatType:
			totalSize += 8 // Double
		case *types.PointerType:
			totalSize += 8 // Pointers are 8 bytes
		case *types.ArrayType:
			totalSize += 8 // Array pointer
		default:
			totalSize += 8 // Default to 8 bytes for other types
		}
	}

	// Allocate memory for the object
	malloc := cg.module.NewFunc("malloc", types.NewPointer(types.I8),
		ir.NewParam("size", types.I64))
	size := constant.NewInt(types.I64, totalSize)
	objPtr := block.NewCall(malloc, size)

	// Cast to correct type
	typedPtr := block.NewBitCast(objPtr, types.NewPointer(classType))

	// Store vtable pointer
	vtablePtr := block.NewGetElementPtr(info.vtableType, cg.vtables[className],
		constant.NewInt(types.I32, 0))
	block.NewStore(vtablePtr,
		block.NewGetElementPtr(classType, typedPtr,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, 0)))

	// Initialize attributes with default values
	for i, attrType := range info.attributeTypes {
		attrPtr := block.NewGetElementPtr(classType, typedPtr,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, int64(i+1)))

		var defaultValue value.Value
		switch t := attrType.(type) {
		case *types.IntType:
			defaultValue = constant.NewInt(types.I64, 0)
		case *types.FloatType:
			defaultValue = constant.NewFloat(types.Double, 0.0)
		case *types.PointerType:
			defaultValue = constant.NewNull(t)
		case *types.ArrayType:
			defaultValue = constant.NewNull(types.NewPointer(t))
		default:
			defaultValue = constant.NewNull(types.NewPointer(types.I8))
		}
		block.NewStore(defaultValue, attrPtr)
	}

	return typedPtr, block, nil
}
