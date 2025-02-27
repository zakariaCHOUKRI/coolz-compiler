package codegen

import (
	"coolz-compiler/ast"
	"fmt"
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
	malloc          *ir.Func
	memcpy          *ir.Func
	currentBindings map[string]value.Value
	currentTypes    map[string]string // Add this map to store variable -> COOL type
	blockCounter    int
	classParents    map[string]string
	classLayouts    map[string]*types.StructType
	classFields     map[string]map[string]int // Maps class->field->index
	currentClass    string
	strlen          *ir.Func
	program         *ast.Program // Add this field
}

// In the New() function, add malloc declaration:
func New() *CodeGenerator {
	cg := &CodeGenerator{
		module:          ir.NewModule(),
		stringConstants: make(map[string]*ir.Global),
		methods:         make(map[string]map[string]*ir.Func),
		currentBindings: make(map[string]value.Value),
		currentTypes:    make(map[string]string),
		classParents:    make(map[string]string),
		classLayouts:    make(map[string]*types.StructType),
		classFields:     make(map[string]map[string]int),
	}

	// Declare external functions
	printfType := types.NewPointer(types.I8)
	cg.printf = cg.module.NewFunc("printf", types.I32, ir.NewParam("format", printfType))
	cg.printf.Sig.Variadic = true

	scanfType := types.NewPointer(types.I8)
	cg.scanf = cg.module.NewFunc("scanf", types.I32, ir.NewParam("format", scanfType))
	cg.scanf.Sig.Variadic = true

	memsetType := types.NewPointer(types.I8)
	cg.memset = cg.module.NewFunc("memset", types.NewPointer(types.I8),
		ir.NewParam("str", memsetType),
		ir.NewParam("c", types.I32),
		ir.NewParam("n", types.I64))

	// Add malloc declaration
	cg.malloc = cg.module.NewFunc("malloc", types.NewPointer(types.I8),
		ir.NewParam("size", types.I64))

	// Add memcpy declaration
	cg.memcpy = cg.module.NewFunc("memcpy", types.NewPointer(types.I8),
		ir.NewParam("dest", types.NewPointer(types.I8)),
		ir.NewParam("src", types.NewPointer(types.I8)),
		ir.NewParam("size", types.I64))

	cg.strlen = cg.module.NewFunc("strlen", types.I64,
		ir.NewParam("str", types.NewPointer(types.I8)))

	return cg
}

// Generate generates LLVM IR for the entire program
func (cg *CodeGenerator) Generate(program *ast.Program) (*ir.Module, error) {
	cg.program = program

	// Initialize Object class as the root
	cg.classParents["Object"] = ""
	cg.methods["Object"] = make(map[string]*ir.Func)

	// Add abort() method
	abortFunc := cg.module.NewFunc("Object_abort", types.NewPointer(types.I8),
		ir.NewParam("self", types.NewPointer(types.I8)))
	block := abortFunc.NewBlock("")

	// Print "abort\n" and exit
	abortStr := cg.getStringConstant("Error: the program was aborted by an abort() function\n")
	block.NewCall(cg.printf, abortStr)
	exitFunc := cg.module.NewFunc("exit", types.Void,
		ir.NewParam("status", types.I32))
	block.NewCall(exitFunc, constant.NewInt(types.I32, 1))
	block.NewUnreachable()
	cg.methods["Object"]["abort"] = abortFunc

	// Modify the type_name() method initialization
	typeNameFunc := cg.module.NewFunc("Object_type_name", types.NewPointer(types.I8),
		ir.NewParam("self", types.NewPointer(types.I8)))
	block = typeNameFunc.NewBlock("")

	// Now we'll set the return type of type_name to be a String
	objStr := cg.getStringConstant("Object")
	block.NewRet(objStr)
	cg.methods["Object"]["type_name"] = typeNameFunc

	// Important: Add type_name to return String type in our type tracking
	for className := range cg.methods {
		if cg.currentTypes == nil {
			cg.currentTypes = make(map[string]string)
		}
		cg.currentTypes[className+"_type_name"] = "String"
	}

	// Add copy() method
	copyFunc := cg.module.NewFunc("Object_copy", types.NewPointer(types.I8),
		ir.NewParam("self", types.NewPointer(types.I8)))
	block = copyFunc.NewBlock("")
	// Create a shallow copy
	layout := cg.classLayouts["Object"] // or any needed layout
	// Rough size calculation (not strictly correct for all alignments, but illustrative)
	cg.classLayouts["Object"] = types.NewStruct(
		types.NewPointer(types.I8), // vtable pointer
	)
	structSize := cg.calculateStructSize(layout)
	newObj := block.NewCall(cg.malloc, constant.NewInt(types.I64, structSize))
	block.NewCall(cg.memcpy, newObj, copyFunc.Params[0], constant.NewInt(types.I64, structSize))
	block.NewRet(newObj)
	cg.methods["Object"]["copy"] = copyFunc

	// Initialize IO class methods
	cg.classParents["IO"] = "Object" // IO inherits from Object
	cg.methods["IO"] = make(map[string]*ir.Func)

	// Create out_string method
	outString := cg.module.NewFunc("IO_out_string", types.NewPointer(types.I8),
		ir.NewParam("self", types.NewPointer(types.I8)),
		ir.NewParam("x", types.NewPointer(types.I8)))
	cg.methods["IO"]["out_string"] = outString

	block = outString.NewBlock("")

	// Create format string for printing strings
	strFormat := cg.getStringConstant("%s")

	// Call printf with the string
	block.NewCall(cg.printf, strFormat, outString.Params[1])
	block.NewRet(outString.Params[0])

	// Create out_int method
	outInt := cg.module.NewFunc("IO_out_int", types.NewPointer(types.I8),
		ir.NewParam("self", types.NewPointer(types.I8)),
		ir.NewParam("x", types.I64))
	cg.methods["IO"]["out_int"] = outInt

	block = outInt.NewBlock("")

	// Create format string for printing integers
	intFormat := cg.getStringConstant("%lld")

	// Call printf with the integer
	block.NewCall(cg.printf, intFormat, outInt.Params[1])
	block.NewRet(outInt.Params[0])

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

	// Clear the newline from input buffer
	block.NewCall(cg.scanf, cg.getStringConstant("%*c"))

	strPtr = block.NewGetElementPtr(types.NewArray(256, types.I8), buffer, zero, zero)
	length := block.NewCall(cg.strlen, strPtr)

	// Allocate permanent storage for the string (+1 for null terminator)
	size := block.NewAdd(length, constant.NewInt(types.I64, 1))
	permanent := block.NewCall(cg.malloc, size)

	// Copy the string to permanent storage
	block.NewCall(cg.memcpy, permanent, strPtr, size)

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

	// Initialize String class methods
	cg.classParents["String"] = "Object" // String inherits from Object
	cg.methods["String"] = make(map[string]*ir.Func)

	// Create length() method
	lengthFunc := cg.module.NewFunc("String_length", types.I64,
		ir.NewParam("self", types.NewPointer(types.I8)))
	cg.methods["String"]["length"] = lengthFunc

	block = lengthFunc.NewBlock("")
	callResult := block.NewCall(cg.strlen, lengthFunc.Params[0])
	block.NewRet(callResult)

	// Create substr() method
	substrFunc := cg.module.NewFunc("String_substr", types.NewPointer(types.I8),
		ir.NewParam("self", types.NewPointer(types.I8)),
		ir.NewParam("i", types.I64),
		ir.NewParam("l", types.I64))
	cg.methods["String"]["substr"] = substrFunc

	block = substrFunc.NewBlock("")

	// Get string length for bounds checking
	strLen := block.NewCall(cg.strlen, substrFunc.Params[0])

	// Check if i is negative
	iNegative := block.NewICmp(enum.IPredSLT, substrFunc.Params[1], constant.NewInt(types.I64, 0))

	// Check if i + l is greater than string length
	iSum := block.NewAdd(substrFunc.Params[1], substrFunc.Params[2])
	outOfBounds := block.NewICmp(enum.IPredSGT, iSum, strLen)

	// Check if l is negative
	lNegative := block.NewICmp(enum.IPredSLT, substrFunc.Params[2], constant.NewInt(types.I64, 0))

	// Combine all error conditions
	errorCond := block.NewOr(iNegative, outOfBounds)
	errorCond = block.NewOr(errorCond, lNegative)

	// Create blocks for error and success paths
	errorBlock := substrFunc.NewBlock("error")
	successBlock := substrFunc.NewBlock("success")

	// Branch based on error condition
	block.NewCondBr(errorCond, errorBlock, successBlock)

	// Error block: print error message and abort
	errorBlock.NewCall(cg.printf,
		cg.getStringConstant("Error: substr out of range\n"))
	abortFunc = cg.methods["Object"]["abort"] // Changed from := to =
	errorBlock.NewCall(abortFunc, substrFunc.Params[0])
	errorBlock.NewUnreachable()

	// Success block: create the substring
	// Get pointer to start of substring
	startPtr := successBlock.NewGetElementPtr(types.I8, substrFunc.Params[0], substrFunc.Params[1])

	// Allocate memory for new string (+1 for null terminator)
	substrSize := successBlock.NewAdd(substrFunc.Params[2], constant.NewInt(types.I64, 1))
	newStr := successBlock.NewCall(cg.malloc, substrSize)

	// Copy the substring
	successBlock.NewCall(cg.memcpy, newStr, startPtr, substrFunc.Params[2])

	// Add null terminator
	endPtr := successBlock.NewGetElementPtr(types.I8, newStr, substrFunc.Params[2])
	successBlock.NewStore(constant.NewInt(types.I8, 0), endPtr)

	// Return the new string
	successBlock.NewRet(newStr)

	// Create concat() method
	concatFunc := cg.module.NewFunc("String_concat", types.NewPointer(types.I8),
		ir.NewParam("self", types.NewPointer(types.I8)),
		ir.NewParam("s", types.NewPointer(types.I8)))
	cg.methods["String"]["concat"] = concatFunc

	block = concatFunc.NewBlock("")

	// Get lengths of both strings
	selfLen := block.NewCall(cg.strlen, concatFunc.Params[0])
	sLen := block.NewCall(cg.strlen, concatFunc.Params[1])

	// Calculate total length needed (+1 for null terminator)
	totalLen := block.NewAdd(selfLen, sLen)
	allocSize := block.NewAdd(totalLen, constant.NewInt(types.I64, 1))

	// Allocate memory for new string
	newStr2 := block.NewCall(cg.malloc, allocSize)

	// Copy first string (self)
	block.NewCall(cg.memcpy, newStr2, concatFunc.Params[0], selfLen)

	// Calculate pointer to where second string should go
	secondStrPtr := block.NewGetElementPtr(types.I8, newStr2, selfLen)

	// Copy second string (including null terminator)
	sLenPlusOne := block.NewAdd(sLen, constant.NewInt(types.I64, 1))
	block.NewCall(cg.memcpy, secondStrPtr, concatFunc.Params[1], sLenPlusOne)

	// Return the concatenated string
	block.NewRet(newStr2)

	// First pass: Register all classes, inheritance, and methods
	for _, class := range program.Classes {
		className := class.Name.Value

		// Register inheritance
		if class.Parent != nil {
			cg.classParents[className] = class.Parent.Value
		} else if className != "Object" {
			cg.classParents[className] = "Object"
		}

		// Create class layout
		cg.createClassLayout(className, program)

		// Register methods (including inherited ones)
		err := cg.registerClassMethods(class)
		if err != nil {
			return nil, err
		}
	}

	// Second pass: Generate all class methods and bodies
	for _, class := range program.Classes {
		err := cg.generateClass(class, program)
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

// Add this helper function
func (cg *CodeGenerator) registerClassMethods(class *ast.Class) error {
	className := class.Name.Value

	// Initialize method map if not exists
	if cg.methods[className] == nil {
		cg.methods[className] = make(map[string]*ir.Func)
	}

	// If this class has a parent, inherit its methods first
	if class.Parent != nil {
		parentName := class.Parent.Value
		if parentMethods, exists := cg.methods[parentName]; exists {
			// Copy parent methods
			for methodName, method := range parentMethods {
				// Only copy if the method isn't overridden in current class
				if _, overridden := cg.methods[className][methodName]; !overridden {
					cg.methods[className][methodName] = method
				}
			}
		}
	}

	// Register class's own methods
	for _, feature := range class.Features {
		if method, ok := feature.(*ast.Method); ok {
			// Create function parameters
			params := make([]*ir.Param, 0, len(method.Formals)+1)
			params = append(params, ir.NewParam("self", types.NewPointer(types.I8)))

			for _, formal := range method.Formals {
				paramType := cg.getLLVMType(formal.Type)
				params = append(params, ir.NewParam(formal.Name.Value, paramType))
			}

			// Create function
			returnType := cg.getLLVMType(method.Type)
			fn := cg.module.NewFunc(fmt.Sprintf("%s_%s", className, method.Name.Value),
				returnType, params...)

			// Register the method
			cg.methods[className][method.Name.Value] = fn
		}
	}

	return nil
}

func (cg *CodeGenerator) calculateStructSize(layout *types.StructType) int64 {
	if layout == nil {
		return 8 // Return minimum size for nil layouts (pointer size)
	}

	structSize := int64(0)
	for _, field := range layout.Fields {
		switch ft := field.(type) {
		case *types.IntType:
			structSize += int64(ft.BitSize) / 8
		case *types.PointerType:
			structSize += 8
		case *types.ArrayType:
			structSize += int64(ft.Len)
		case *types.StructType:
			structSize += cg.calculateStructSize(ft)
		default:
			structSize += 8
		}
	}
	return ((structSize + 7) / 8) * 8
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

func (cg *CodeGenerator) generateClass(class *ast.Class, program *ast.Program) error {
	// Save previous class
	prevClass := cg.currentClass
	cg.currentClass = class.Name.Value
	defer func() { cg.currentClass = prevClass }()

	// Create class layout first
	cg.createClassLayout(class.Name.Value, program)

	className := class.Name.Value

	// Override type_name method for this class
	typeNameFunc := cg.module.NewFunc(fmt.Sprintf("%s_type_name", className),
		types.NewPointer(types.I8),
		ir.NewParam("self", types.NewPointer(types.I8)))

	block := typeNameFunc.NewBlock("")
	classNameStr := cg.getStringConstant(className)
	block.NewRet(classNameStr)
	cg.methods[className]["type_name"] = typeNameFunc

	// Second pass: Generate method bodies
	for _, feature := range class.Features {
		if method, ok := feature.(*ast.Method); ok {
			prevFunc := cg.currentFunc
			cg.currentFunc = cg.methods[className][method.Name.Value]

			err := cg.generateMethodBody(className, method)
			if err != nil {
				return err
			}

			cg.currentFunc = prevFunc
		}
	}

	return nil
}

func (cg *CodeGenerator) generateMethodBody(className string, method *ast.Method) error {
	// Save previous state
	prevBindings := cg.currentBindings
	prevTypes := cg.currentTypes
	cg.currentBindings = make(map[string]value.Value)
	cg.currentTypes = make(map[string]string)

	fn := cg.methods[className][method.Name.Value]
	block := fn.NewBlock("")

	// Add class attributes to scope first
	self := fn.Params[0]
	structPtr := block.NewBitCast(self, types.NewPointer(cg.classLayouts[className]))

	// Add all attributes from the class hierarchy
	currentClass := className
	for currentClass != "" {
		if fields, exists := cg.classFields[currentClass]; exists {
			for fieldName, fieldIndex := range fields {
				fieldPtr := block.NewGetElementPtr(cg.classLayouts[className], structPtr,
					constant.NewInt(types.I32, 0),
					constant.NewInt(types.I32, int64(fieldIndex)))
				cg.currentBindings[fieldName] = fieldPtr
				// Store the attribute type
				for _, class := range cg.program.Classes {
					if class.Name.Value == currentClass {
						for _, feature := range class.Features {
							if attr, ok := feature.(*ast.Attribute); ok && attr.Name.Value == fieldName {
								cg.currentTypes[fieldName] = attr.Type.Value
							}
						}
					}
				}
			}
		}
		currentClass = cg.classParents[currentClass]
	}

	// Store parameters in allocas
	for i, formal := range method.Formals {
		alloca := block.NewAlloca(fn.Params[i+1].Type())
		block.NewStore(fn.Params[i+1], alloca)
		cg.currentBindings[formal.Name.Value] = alloca
		cg.currentTypes[formal.Name.Value] = formal.Type.Value
	}

	value, block, err := cg.generateExpression(block, method.Body)
	if err != nil {
		return err
	}

	if block.Term == nil {
		block.NewRet(value)
	}

	// Restore previous state
	cg.currentBindings = prevBindings
	cg.currentTypes = prevTypes
	return nil
}

// Walk upward through parents until we find the method or run out of parents
func (cg *CodeGenerator) lookupMethod(className, methodName string) (*ir.Func, bool) {
	currentClass := className
	for currentClass != "" {
		if methods, exists := cg.methods[currentClass]; exists {
			if method, exists := methods[methodName]; exists {
				return method, true
			}
		}
		currentClass = cg.classParents[currentClass]
	}
	return nil, false
}

func getMethodNames(methods map[string]*ir.Func) []string {
	names := make([]string, 0, len(methods))
	for name := range methods {
		names = append(names, name)
	}
	return names
}

func (cg *CodeGenerator) generateMethodCall(block *ir.Block, object value.Value, className string,
	methodName string, args []ast.Expression) (value.Value, *ir.Block, error) {

	// Look up the method in the class hierarchy
	method, exists := cg.lookupMethod(className, methodName)
	if !exists {
		return nil, block, fmt.Errorf("method %s not found in class %s or its parents", methodName, className)
	}

	// Generate code for each argument
	llvmArgs := []value.Value{object} // First argument is always the object itself
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

// generateExpression now returns (value, currentBlock, error)
// so that expressions which change control flow (like if) can update the current block.
func (cg *CodeGenerator) generateExpression(block *ir.Block, expr ast.Expression) (value.Value, *ir.Block, error) {
	switch e := expr.(type) {
	case *ast.BooleanLiteral:
		// In LLVM, booleans are represented as i1 (1-bit integers)
		// true = 1, false = 0
		if e.Value {
			return constant.NewInt(types.I1, 1), block, nil
		}
		return constant.NewInt(types.I1, 0), block, nil
	case *ast.StringLiteral:
		return cg.getStringConstant(e.Value), block, nil
	case *ast.IntegerLiteral:
		return constant.NewInt(types.I64, e.Value), block, nil
	case *ast.Self:
		// In COOL, 'self' refers to the first parameter of every method
		// which is always the current object instance
		if cg.currentFunc == nil {
			return nil, block, fmt.Errorf("self used outside of method context")
		}
		// Return the first parameter (self) of the current function
		return cg.currentFunc.Params[0], block, nil
		// In the DynamicDispatch case in generateExpression, modify the type determination section:
	case *ast.DynamicDispatch:
		methodName := e.Method.Value

		// First, generate code for the object we're dispatching on
		var objValue value.Value
		var objType string

		if e.Object == nil || e.Object.TokenLiteral() == "self" {
			// Self dispatch handling
			if cg.currentFunc == nil {
				return nil, block, fmt.Errorf("dispatch on self outside method context")
			}
			objValue = cg.currentFunc.Params[0]
			// Use the current class type if we're in a method
			objType = strings.Split(cg.currentFunc.Name(), "_")[0]
		} else {
			// Generate code for the object expression
			var newBlock *ir.Block
			var err error
			objValue, newBlock, err = cg.generateExpression(block, e.Object)
			if err != nil {
				return nil, block, err
			}
			block = newBlock

			// Determine the type of the object
			switch obj := e.Object.(type) {
			case *ast.ObjectIdentifier:
				// Look up the type in our current types map
				var exists bool
				objType, exists = cg.currentTypes[obj.Value]
				if !exists {
					return nil, block, fmt.Errorf("undefined variable: %s", obj.Value)
				}
			case *ast.NewExpression:
				objType = obj.Type.Value
			case *ast.DynamicDispatch:
				// For nested dispatches, first determine the type of the method result
				if objType == "" {
					// If objType is not set, we need to determine it from the nested dispatch
					switch obj.Object.(type) {
					case *ast.StringLiteral:
						objType = "String"
					case *ast.IntegerLiteral:
						objType = "Int"
					case *ast.BooleanLiteral:
						objType = "Bool"
					default:
						// For other cases, get the type from current types map
						if innerObjId, ok := obj.Object.(*ast.ObjectIdentifier); ok {
							if t, exists := cg.currentTypes[innerObjId.Value]; exists {
								objType = t
							}
						}
					}
				}

				// Look up the method in the current objType
				if method, exists := cg.lookupMethod(objType, obj.Method.Value); exists {
					// For type_name method, we know it always returns String
					if obj.Method.Value == "type_name" {
						objType = "String"
					} else {
						// For other methods, extract return type from method name
						parts := strings.Split(method.Name(), "_")
						if len(parts) > 0 {
							objType = parts[0] // The class name is always the first part
						}
					}
				} else {
					return nil, block, fmt.Errorf("method %s not found in class %s", obj.Method.Value, objType)
				}

			case *ast.StringLiteral:
				objType = "String"
			case *ast.IntegerLiteral:
				objType = "Int"
			case *ast.BooleanLiteral:
				objType = "Bool"
			default:
				return nil, block, fmt.Errorf("unsupported dispatch object type: %T", e.Object)
			}
		}

		// Now generate the method call with the correct object and type
		return cg.generateMethodCall(block, objValue, objType, methodName, e.Arguments)
	case *ast.BlockExpression:
		return cg.generateBlock(block, e)
	case *ast.LetExpression:
		return cg.generateLet(block, e)
	case *ast.ObjectIdentifier:
		// First check if it's a local variable
		if alloca, exists := cg.currentBindings[e.Value]; exists {
			// If it's a pointer to a field, just return the pointer
			if strings.Contains(alloca.Type().String(), "GEP") {
				return alloca, block, nil
			}
			// Otherwise load the value
			return block.NewLoad(alloca.Type().(*types.PointerType).ElemType, alloca), block, nil
		}

		// If not found as local, check if it's a class attribute
		if cg.currentFunc != nil && cg.currentClass != "" {
			self := cg.currentFunc.Params[0]
			structPtr := block.NewBitCast(self, types.NewPointer(cg.classLayouts[cg.currentClass]))

			// Look up through the class hierarchy
			currentClass := cg.currentClass
			for currentClass != "" {
				if fieldIndex, exists := cg.classFields[currentClass][e.Value]; exists {
					fieldPtr := block.NewGetElementPtr(cg.classLayouts[cg.currentClass], structPtr,
						constant.NewInt(types.I32, 0),
						constant.NewInt(types.I32, int64(fieldIndex)))

					// Load and return the field value
					fieldType := cg.classLayouts[cg.currentClass].Fields[fieldIndex]
					return block.NewLoad(fieldType, fieldPtr), block, nil
				}
				currentClass = cg.classParents[currentClass]
			}
		}

		return nil, block, fmt.Errorf("undefined variable or attribute: %s", e.Value)
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
		return cg.generateAssignment(block, e)
	case *ast.NewExpression:
		// Placeholder for new object creation
		newObj := constant.NewNull(types.NewPointer(types.I8))
		return newObj, block, nil
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
	prevBindings := make(map[string]value.Value)
	prevTypes := make(map[string]string)

	// Save old bindings and types
	for k, v := range cg.currentBindings {
		prevBindings[k] = v
	}
	for k, v := range cg.currentTypes {
		prevTypes[k] = v
	}

	currentBlock := block
	for _, binding := range letExpr.Bindings {
		// uniqueName := fmt.Sprintf("%s_let%d_%d", binding.Identifier.Value, len(cg.currentBindings), i)
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

		// Store the alloca and the COOL type
		cg.currentBindings[binding.Identifier.Value] = alloca
		cg.currentTypes[binding.Identifier.Value] = binding.Type.Value
	}

	result, newBlock, err := cg.generateExpression(currentBlock, letExpr.In)

	// Restore old bindings and types
	cg.currentBindings = prevBindings
	cg.currentTypes = prevTypes

	return result, newBlock, err
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

func (cg *CodeGenerator) createClassLayout(className string, program *ast.Program) *types.StructType {
	if layout, exists := cg.classLayouts[className]; exists {
		return layout
	}

	// Get parent class layout first
	var fields []types.Type
	if parent, exists := cg.classParents[className]; exists && parent != "" {
		parentLayout := cg.createClassLayout(parent, program)
		fields = append(fields, parentLayout.Fields...)
	}

	// Add vtable pointer (for methods)
	fields = append(fields, types.NewPointer(types.I8))

	// Create field map if it doesn't exist
	if cg.classFields[className] == nil {
		cg.classFields[className] = make(map[string]int)
	}

	// Add class's own fields
	for _, class := range program.Classes {
		if class.Name.Value == className {
			for _, f := range class.Features {
				if attr, ok := f.(*ast.Attribute); ok {
					cg.classFields[className][attr.Name.Value] = len(fields)
					fields = append(fields, cg.getLLVMType(attr.Type))
				}
			}
			break
		}
	}

	layout := types.NewStruct(fields...)
	cg.classLayouts[className] = layout
	return layout
}

// Now generateAssignment can use cg.currentClass
func (cg *CodeGenerator) generateAssignment(block *ir.Block, assign *ast.Assignment) (value.Value, *ir.Block, error) {
	if obj, ok := assign.Left.(*ast.ObjectIdentifier); ok {
		// First check if this is a field access
		if fieldIndex, exists := cg.classFields[cg.currentClass][obj.Value]; exists {
			self := cg.currentFunc.Params[0] // get self parameter
			// Cast self to struct pointer
			structPtr := block.NewBitCast(self, types.NewPointer(cg.classLayouts[cg.currentClass]))
			// Generate field pointer
			fieldPtr := block.NewGetElementPtr(cg.classLayouts[cg.currentClass], structPtr,
				constant.NewInt(types.I32, 0),
				constant.NewInt(types.I32, int64(fieldIndex)))

			// Generate value and store it
			value, newBlock, err := cg.generateExpression(block, assign.Value)
			if err != nil {
				return nil, block, err
			}
			newBlock.NewStore(value, fieldPtr)
			return value, newBlock, nil
		}

		// If not a field, check local variables
		if alloca, exists := cg.currentBindings[obj.Value]; exists {
			// Generate code for the value expression
			value, newBlock, err := cg.generateExpression(block, assign.Value)
			if err != nil {
				return nil, block, err
			}
			// Store the new value
			newBlock.NewStore(value, alloca)
			return value, newBlock, nil
		}
	}
	return nil, block, fmt.Errorf("undefined variable or field: %v", assign.Left)
}
