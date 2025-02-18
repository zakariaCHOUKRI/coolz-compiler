package codegen

import (
	"bytes"
	"coolz-compiler/ast"
	"fmt"
	"strings"
)

type CodeGenerator struct {
	buf     bytes.Buffer
	strings map[string]string // Tracks string literals and their global names
	formats map[string]string // Tracks printf format strings
}

func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		strings: make(map[string]string),
		formats: make(map[string]string),
	}
}

func (cg *CodeGenerator) Generate(program *ast.Program) string {
	cg.buf.Reset()
	cg.emitHeader()

	// Generate code for classes first
	for _, class := range program.Classes {
		cg.genClass(class)
	}

	// Emit global strings afterward
	for value, name := range cg.strings {
		cg.emitGlobalString(name, value)
	}
	// Handle format strings separately
	for format, name := range cg.formats {
		// For format strings, we need to account for the format specifier
		cg.emit("%s = private unnamed_addr constant [4 x i8] c\"%s\\00\\00\"\n", name, format)
	}

	cg.emitMainFunction()
	return cg.buf.String()
}

func (cg *CodeGenerator) emitHeader() {
	cg.emit("declare i32 @printf(i8*, ...)\n\n")
}

func (cg *CodeGenerator) genClass(class *ast.Class) {
	for _, feature := range class.Features {
		switch f := feature.(type) {
		case *ast.Method:
			cg.genMethod(class.Name.Value, f)
		}
	}
}

func (cg *CodeGenerator) genMethod(className string, method *ast.Method) {
	funcName := fmt.Sprintf("%s_%s", className, method.Name.Value)
	cg.emit("define i8* @%s(i8* %%self) {\n", funcName)
	cg.genExpression(method.Body)
	cg.emit("ret i8* %%self\n}\n\n")
}

func (cg *CodeGenerator) genExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.BlockExpression:
		for _, exp := range e.Expressions {
			cg.genExpression(exp)
		}
	case *ast.DynamicDispatch:
		cg.genDynamicDispatch(e)
	case *ast.StringLiteral:
		// String literals are handled during their usage in dispatch
	case *ast.IntegerLiteral:
		// Integer literals are handled during their usage in dispatch
	default:
		// Handle other expressions as needed
	}
}

func (cg *CodeGenerator) genDynamicDispatch(d *ast.DynamicDispatch) {
	methodName := d.Method.Value
	switch methodName {
	case "out_string":
		cg.genOutString(d)
	case "out_int":
		cg.genOutInt(d)
	// Add cases for in_string, in_int, etc. as needed
	default:
		// Handle other methods
	}
}

func (cg *CodeGenerator) genOutString(d *ast.DynamicDispatch) {
	if len(d.Arguments) != 1 {
		panic("out_string requires exactly one argument")
	}

	strLit, ok := d.Arguments[0].(*ast.StringLiteral)
	if !ok {
		panic("out_string argument must be a string literal")
	}

	// Generate string constant and format
	strName := cg.genStringConstant(strLit.Value)
	formatName := cg.genFormatString("%s")

	cg.emit("  %%1 = call i32 (i8*, ...) @printf(i8* getelementptr inbounds ([3 x i8], [3 x i8]* %s, i32 0, i32 0), i8* getelementptr inbounds ([%d x i8], [%d x i8]* %s, i32 0, i32 0))\n",
		formatName, len(strLit.Value)+1, len(strLit.Value)+1, strName)
}

func (cg *CodeGenerator) genOutInt(d *ast.DynamicDispatch) {
	if len(d.Arguments) != 1 {
		panic("out_int requires exactly one argument")
	}

	intLit, ok := d.Arguments[0].(*ast.IntegerLiteral)
	if !ok {
		panic("out_int argument must be an integer literal")
	}

	formatName := cg.genFormatString("%d")
	cg.emit("  %%1 = call i32 (i8*, ...) @printf(i8* getelementptr inbounds ([3 x i8], [3 x i8]* %s, i32 0, i32 0), i32 %d)\n",
		formatName, intLit.Value)
}

func (cg *CodeGenerator) genStringConstant(value string) string {
	if name, exists := cg.strings[value]; exists {
		return name
	}
	name := fmt.Sprintf("@.str.%d", len(cg.strings))
	cg.strings[value] = name
	return name
}

func (cg *CodeGenerator) genFormatString(format string) string {
	if name, exists := cg.formats[format]; exists {
		return name
	}
	safeFormat := strings.ReplaceAll(format, "%", "p")
	name := fmt.Sprintf("@.fmt.%s", safeFormat)
	cg.formats[format] = name
	return name
}

func (cg *CodeGenerator) emitGlobalString(name, value string) {
	escaped := strings.ReplaceAll(value, "\n", "\\0A")
	escaped = strings.ReplaceAll(escaped, "\"", "\\22")
	actualLength := len(value) + 1 // Original string length + null terminator
	cg.emit("%s = private unnamed_addr constant [%d x i8] c\"%s\\00\"\n", name, actualLength, escaped)
}

func (cg *CodeGenerator) emitMainFunction() {
	cg.emit("define i32 @main() {\n")
	cg.emit("  %%1 = call i8* @Main_main(i8* null)\n")
	cg.emit("  ret i32 0\n}\n")
}

func (cg *CodeGenerator) emit(format string, args ...interface{}) {
	cg.buf.WriteString(fmt.Sprintf(format, args...))
}
