package irgen

import (
	"coolz-compiler/ast"
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum" // Add this import
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

type IRGenerator struct {
	module       *ir.Module
	currentBlock *ir.Block
	currentFunc  *ir.Func
	symbolTable  map[string]value.Value
	classTypes   map[string]*types.StructType
	classVTables map[string]*ir.Global
	stringCache  map[string]*ir.Global
}

// Remove builder field and its initialization since llir/llvm doesn't have a Builder

func NewIRGenerator() *IRGenerator {
	return &IRGenerator{
		module:       ir.NewModule(),
		symbolTable:  make(map[string]value.Value),
		classTypes:   make(map[string]*types.StructType),
		classVTables: make(map[string]*ir.Global),
		stringCache:  make(map[string]*ir.Global),
	}
}

func (g *IRGenerator) Generate(program *ast.Program) (*ir.Module, error) {
	// Define basic types
	g.defineBuiltinTypes()

	// Generate class types and vtables
	for _, class := range program.Classes {
		g.generateClassType(class)
	}

	// Generate methods
	for _, class := range program.Classes {
		g.generateClassMethods(class)
	}

	// Create main function
	g.generateMainFunction()

	return g.module, nil
}

func (g *IRGenerator) defineBuiltinTypes() {
	// Object type
	vtablePtr := types.I8Ptr
	objType := types.NewStruct(vtablePtr)
	g.classTypes["Object"] = objType
	g.module.NewTypeDef("Object", objType)

	// Int type
	intType := types.NewStruct(
		vtablePtr, // vtable pointer
		types.I32, // value
	)
	g.classTypes["Int"] = intType
	g.module.NewTypeDef("Int", intType)

	// String type
	strType := types.NewStruct(
		vtablePtr, // vtable pointer
		vtablePtr, // char* pointer
		types.I32, // length
	)
	g.classTypes["String"] = strType
	g.module.NewTypeDef("String", strType)

	// Bool type
	boolType := types.NewStruct(
		vtablePtr, // vtable pointer
		types.I1,  // value
	)
	g.classTypes["Bool"] = boolType
	g.module.NewTypeDef("Bool", boolType)
}

func (g *IRGenerator) generateClassType(class *ast.Class) error {
	className := class.Name.Value

	// Create struct type for class
	fields := []types.Type{
		types.I8Ptr, // vtable pointer
	}

	// Add fields for attributes
	for _, feature := range class.Features {
		if attr, ok := feature.(*ast.Attribute); ok {
			if typ := g.typeFromString(attr.Type.Value); typ != nil {
				fields = append(fields, typ)
			}
		}
	}

	structType := types.NewStruct(fields...)
	g.classTypes[className] = structType
	g.module.NewTypeDef(className, structType)

	return nil
}

func (g *IRGenerator) generateClassMethods(class *ast.Class) error {
	className := class.Name.Value

	for _, feature := range class.Features {
		if method, ok := feature.(*ast.Method); ok {
			returnTyp := g.typeFromString(method.Type.Value)

			// Create parameters
			params := []*ir.Param{
				ir.NewParam("self", types.NewPointer(g.classTypes[className])),
			}

			// Add method parameters
			for _, formal := range method.Formals {
				paramType := g.typeFromString(formal.Type.Value)
				params = append(params, ir.NewParam(formal.Name.Value, paramType))
			}

			// Create function
			funcName := fmt.Sprintf("%s_%s", className, method.Name.Value)
			f := g.module.NewFunc(funcName, returnTyp, params...)

			// Generate method body
			g.currentFunc = f
			g.currentBlock = f.NewBlock("")

			result := g.generateExpression(method.Body)
			g.currentBlock.NewRet(result)
		}
	}

	return nil
}

func (g *IRGenerator) generateMainFunction() error {
	mainType := types.NewPointer(g.classTypes["Object"])
	main := g.module.NewFunc("main", types.I32)
	block := main.NewBlock("")
	g.currentBlock = block

	// Create Main object and call Main.main()
	mainClass := g.getOrDeclareRuntime("Main_new")
	mainObj := block.NewCall(mainClass)
	mainMain := g.module.NewFunc("Main_main", mainType)
	block.NewCall(mainMain, mainObj)

	// Return 0
	block.NewRet(constant.NewInt(types.I32, 0))

	return nil
}

func (g *IRGenerator) typeFromString(typeName string) types.Type {
	if t, ok := g.classTypes[typeName]; ok {
		return types.NewPointer(t)
	}
	// Default to Object type if unknown
	return types.NewPointer(g.classTypes["Object"])
}

func (g *IRGenerator) generateExpression(expr ast.Expression) value.Value {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return g.generateIntegerLiteral(e)
	case *ast.StringLiteral:
		return g.generateStringLiteral(e)
	default:
		return constant.NewNull(types.NewPointer(g.classTypes["Object"]))
	}
}

func (g *IRGenerator) generateIntegerLiteral(il *ast.IntegerLiteral) value.Value {
	// Create new Int instance
	intNew := g.getOrDeclareRuntime("Int_new")
	intObj := g.currentBlock.NewCall(intNew)

	// Set the value
	intPtr := g.currentBlock.NewBitCast(intObj, types.NewPointer(g.classTypes["Int"]))

	// Create GEP indices for accessing the value field (index 1) of the Int struct
	indices := []value.Value{
		constant.NewInt(types.I32, 0), // First index for pointer
		constant.NewInt(types.I32, 1), // Second index for struct field
	}

	// Get pointer to value field
	valuePtr := g.currentBlock.NewGetElementPtr(g.classTypes["Int"], intPtr, indices...)

	// Store the actual integer value
	g.currentBlock.NewStore(constant.NewInt(types.I32, int64(il.Value)), valuePtr)

	return intObj
}

func (g *IRGenerator) generateStringLiteral(sl *ast.StringLiteral) value.Value {
	// Create or get global string constant
	strContent := sl.Value
	var strConst *ir.Global
	if cached, ok := g.stringCache[strContent]; ok {
		strConst = cached
	} else {
		// Create new global string constant
		strConst = g.module.NewGlobalDef(
			fmt.Sprintf(".str.%d", len(g.stringCache)),
			constant.NewCharArrayFromString(strContent+"\x00"),
		)
		g.stringCache[strContent] = strConst
	}

	// Create new String instance
	strNew := g.getOrDeclareRuntime("String_new")
	strObj := g.currentBlock.NewCall(strNew)

	// Cast string object to our String type
	strPtr := g.currentBlock.NewBitCast(strObj, types.NewPointer(g.classTypes["String"]))

	// Get pointer to the char* field (index 1)
	indices := []value.Value{
		constant.NewInt(types.I32, 0),
		constant.NewInt(types.I32, 1),
	}
	charPtrPtr := g.currentBlock.NewGetElementPtr(g.classTypes["String"], strPtr, indices...)

	// Store the pointer to our string content
	strContentPtr := g.currentBlock.NewBitCast(strConst, types.I8Ptr)
	g.currentBlock.NewStore(strContentPtr, charPtrPtr)

	// Set the string length
	lenIndices := []value.Value{
		constant.NewInt(types.I32, 0),
		constant.NewInt(types.I32, 2),
	}
	lenPtr := g.currentBlock.NewGetElementPtr(g.classTypes["String"], strPtr, lenIndices...)
	g.currentBlock.NewStore(constant.NewInt(types.I32, int64(len(strContent))), lenPtr)

	return strObj
}

func (g *IRGenerator) getOrDeclareRuntime(name string) *ir.Func {
	// Try to find existing function
	for _, f := range g.module.Funcs {
		if f.Name() == name {
			return f
		}
	}

	// Declare new function
	var f *ir.Func
	switch name {
	case "Int_new":
		f = g.module.NewFunc(name, types.NewPointer(g.classTypes["Int"]))
	case "String_new":
		f = g.module.NewFunc(name, types.NewPointer(g.classTypes["String"]))
		f.Linkage = enum.LinkageExternal // Use proper enum value
	case "Bool_new":
		f = g.module.NewFunc(name, types.NewPointer(g.classTypes["Bool"]))
	default:
		// For class constructors
		if classType, ok := g.classTypes[name[:len(name)-4]]; ok { // Remove "_new" suffix
			f = g.module.NewFunc(name, types.NewPointer(classType))
		}
	}

	return f
}
