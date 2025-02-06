package semant

import (
	"coolz-compiler/ast"
	"coolz-compiler/lexer"
	"fmt"
	"strings"
)

var (
	debugDepth     = 0
	debugCounter   = 0
	maxDebugOutput = 1000
	callCount      = make(map[string]int)
)

func debug(format string, args ...interface{}) {
	if debugCounter > maxDebugOutput {
		panic(fmt.Sprintf("Debug output limit exceeded. Last message: "+format, args...))
	}
	debugCounter++
	indent := strings.Repeat("  ", debugDepth)
	fmt.Printf(indent+format+"\n", args...)
}

type SymbolTable struct {
	symbols map[string]*SymbolEntry
	parent  *SymbolTable
}

type SymbolEntry struct {
	Type     string
	Token    lexer.Token
	AttrType *ast.TypeIdentifier
	Method   *ast.Method
	Scope    *SymbolTable
}

func NewSymbolTable(parent *SymbolTable) *SymbolTable {
	return &SymbolTable{
		symbols: make(map[string]*SymbolEntry),
		parent:  parent,
	}
}

func (st *SymbolTable) AddEntry(name string, entry *SymbolEntry) {
	st.symbols[name] = entry
}

func (st *SymbolTable) Lookup(name string) (*SymbolEntry, bool) {
	entry, ok := st.symbols[name]
	if !ok && st.parent != nil {
		return st.parent.Lookup(name)
	}
	return entry, ok
}

type SemanticAnalyser struct {
	globalSymbolTable *SymbolTable
	errors            []string
	inheritanceGraph  map[string]string // child -> parent
	currentClass      string            // Add current class tracking
}

func NewSemanticAnalyser() *SemanticAnalyser {
	return &SemanticAnalyser{
		globalSymbolTable: NewSymbolTable(nil),
		errors:            []string{},
		inheritanceGraph:  make(map[string]string),
		currentClass:      "Object",
	}
}

// Add method to resolve SELF_TYPE
func (sa *SemanticAnalyser) resolveSelfType(typeStr string) string {
	if typeStr == "SELF_TYPE" {
		return sa.currentClass
	}
	return typeStr
}

func (sa *SemanticAnalyser) Errors() []string {
	return sa.errors
}

func (sa *SemanticAnalyser) Analyze(program *ast.Program) {
	debug("Starting analysis")
	callCount = make(map[string]int)

	debug("Building inheritance graph")
	sa.buildInheritanceGraph(program)

	debug("Checking inheritance cycles")
	if sa.detectInheritanceCycles() {
		debug("Inheritance cycles detected, stopping analysis")
		return
	}

	debug("Building class symbol tables")
	sa.buildClassesSymboltables(program)

	debug("Building symbol tables")
	sa.buildSymboltables(program)

	debug("Starting type checking")
	sa.typeCheck(program)
}

func (sa *SemanticAnalyser) buildInheritanceGraph(program *ast.Program) {
	// Add basic classes
	sa.inheritanceGraph["Int"] = "Object"
	sa.inheritanceGraph["String"] = "Object"
	sa.inheritanceGraph["Bool"] = "Object"

	for _, class := range program.Classes {
		childName := class.Name.Value
		parentName := "Object"
		if class.Parent != nil {
			parentName = class.Parent.Value
			// Check for invalid inheritance
			if parentName == "Int" || parentName == "String" || parentName == "Bool" {
				sa.errors = append(sa.errors, fmt.Sprintf("Class %s cannot inherit from %s", childName, parentName))
				continue
			}
		}
		sa.inheritanceGraph[childName] = parentName
	}
}

func (sa *SemanticAnalyser) detectInheritanceCycles() bool {
	visited := make(map[string]bool)
	inStack := make(map[string]bool)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		inStack[node] = true

		if parent, ok := sa.inheritanceGraph[node]; ok {
			if !visited[parent] {
				if dfs(parent) {
					return true
				}
			} else if inStack[parent] {
				sa.errors = append(sa.errors, fmt.Sprintf("Inheritance cycle detected involving class %s", node))
				return true
			}
		}

		inStack[node] = false
		return false
	}

	for class := range sa.inheritanceGraph {
		if !visited[class] {
			if dfs(class) {
				return true
			}
		}
	}
	return false
}

func (sa *SemanticAnalyser) typeCheck(program *ast.Program) {
	visited := make(map[string]bool)

	debug("Type checking classes:")
	for _, class := range program.Classes {
		callCount["typeCheck_"+class.Name.Value]++
		if callCount["typeCheck_"+class.Name.Value] > 100 {
			panic(fmt.Sprintf("Potential infinite recursion detected in typeCheck for class %s", class.Name.Value))
		}

		debug("- Checking class: %s", class.Name.Value)
		if !visited[class.Name.Value] {
			visited[class.Name.Value] = true
			if st, ok := sa.globalSymbolTable.symbols[class.Name.Value]; ok && st.Scope != nil {
				sa.typeCheckClass(class, st.Scope)
			} else {
				debug("Warning: No symbol table for class %s", class.Name.Value)
			}
		}
	}
}

// Update typeCheckClass to track current class
func (sa *SemanticAnalyser) typeCheckClass(cls *ast.Class, st *SymbolTable) {
	prevClass := sa.currentClass
	sa.currentClass = cls.Name.Value

	for _, feature := range cls.Features {
		switch f := feature.(type) {
		case *ast.Attribute:
			sa.typeCheckAttribute(f, st)
		case *ast.Method:
			sa.typeCheckMethod(f, st)
		}
	}

	sa.currentClass = prevClass
}

func (sa *SemanticAnalyser) typeCheckAttribute(attribute *ast.Attribute, st *SymbolTable) {
	// Add guard against nil initialization
	if attribute == nil || attribute.Type == nil {
		return
	}

	if attribute.Init != nil {
		expressionType := sa.getExpressionType(attribute.Init, st)
		if expressionType != "Object" && !sa.isTypeConformant(expressionType, attribute.Type.Value) {
			sa.errors = append(sa.errors, fmt.Sprintf("attribute %s cannot be of type %s, expected %s",
				attribute.Name.Value, expressionType, attribute.Type.Value))
		}
	}

}

func (sa *SemanticAnalyser) typeCheckMethod(method *ast.Method, st *SymbolTable) {
	methodSt := st.symbols[method.Name.Value].Scope
	for _, formal := range method.Formals {
		if _, ok := methodSt.Lookup(formal.Name.Value); ok {
			sa.errors = append(sa.errors, fmt.Sprintf("argument %s in method %s is already defined", formal.Name.Value, method.Name.Value))
			continue
		}

		methodSt.parent.AddEntry(formal.Name.Value, &SymbolEntry{Token: method.Token, Type: formal.Type.Value})
	}

	methodExpressionType := sa.getExpressionType(method.Body, methodSt)
	if !sa.isTypeConformant(methodExpressionType, method.Type.Value) {
		sa.errors = append(sa.errors, fmt.Sprintf("method %s is expected to return %s, found %s", method.Name.Value, method.Type.Value, methodExpressionType))
	}
}

// Update isTypeConformant to handle SELF_TYPE properly
func (sa *SemanticAnalyser) isTypeConformant(type1, type2 string) bool {
	debug("Checking type conformance: %s <= %s", type1, type2)
	debugDepth++
	defer func() { debugDepth-- }()

	// Handle SELF_TYPE cases
	if type2 == "SELF_TYPE" {
		return type1 == "SELF_TYPE"
	}

	if type1 == "SELF_TYPE" {
		type1 = sa.currentClass
	}

	if type1 == type2 {
		return true
	}

	// Walk up the inheritance chain
	current := type1
	for {
		parent, ok := sa.inheritanceGraph[current]
		if !ok {
			return false
		}
		if parent == type2 {
			return true
		}
		if parent == "Object" {
			return type2 == "Object"
		}
		current = parent
	}
}

func (sa *SemanticAnalyser) getLeastCommonAncestor(types []string) string {
	if len(types) == 0 {
		return "Object"
	}
	if len(types) == 1 {
		return types[0]
	}

	// Replace SELF_TYPE with current class
	resolvedTypes := make([]string, len(types))
	for i, t := range types {
		resolvedTypes[i] = sa.resolveSelfType(t)
	}

	// Get path to root for first type
	path1 := sa.getPathToRoot(resolvedTypes[0])
	result := resolvedTypes[0]

	// Compare with paths of other types
	for i := 1; i < len(resolvedTypes); i++ {
		path2 := sa.getPathToRoot(resolvedTypes[i])
		result = sa.findFirstCommonAncestor(path1, path2)
		path1 = sa.getPathToRoot(result)
	}

	return result
}

func (sa *SemanticAnalyser) getPathToRoot(typeName string) []string {
	path := []string{typeName}
	current := typeName

	for {
		parent, ok := sa.inheritanceGraph[current]
		if !ok || parent == "Object" {
			path = append(path, "Object")
			break
		}
		path = append(path, parent)
		current = parent
	}

	return path
}

func (sa *SemanticAnalyser) findFirstCommonAncestor(path1, path2 []string) string {
	// Convert paths to sets for faster lookup
	set1 := make(map[string]bool)
	for _, t := range path1 {
		set1[t] = true
	}

	// Find first type in path2 that appears in path1
	for _, t := range path2 {
		if set1[t] {
			return t
		}
	}

	return "Object"
}

func (sa *SemanticAnalyser) getExpressionType(expression ast.Expression, st *SymbolTable) string {
	if expression == nil || st == nil {
		return "Object"
	}

	// Add visited map to prevent infinite recursion in expression type checking
	visited := make(map[ast.Expression]bool)
	return sa.getExpressionTypeInternal(expression, st, visited)
}

func (sa *SemanticAnalyser) getExpressionTypeInternal(expression ast.Expression, st *SymbolTable, visited map[ast.Expression]bool) string {
	callCount["getExpressionTypeInternal"]++
	if callCount["getExpressionTypeInternal"] > 1000 {
		panic("Potential infinite recursion in getExpressionTypeInternal")
	}

	debugDepth++
	defer func() { debugDepth-- }()

	if visited[expression] {
		debug("Cycle detected in expression type checking")
		return "Object"
	}

	debug("Checking expression type: %T", expression)
	visited[expression] = true

	switch e := expression.(type) {
	case *ast.IntegerLiteral:
		return "Int"
	case *ast.StringLiteral:
		return "String"
	case *ast.BooleanLiteral:
		return "Bool"
	case *ast.BlockExpression:
		return sa.getBlockExpressionType(e, st)
	case *ast.IfExpression:
		return sa.getIfExpressionType(e, st)
	case *ast.WhileExpression:
		return sa.getWhileExpressionType(e, st)
	case *ast.NewExpression:
		return sa.getNewExpressionType(e, st)
	case *ast.LetExpression:
		return sa.getLetExpressionType(e, st)
	case *ast.AssignExpression:
		return sa.getAssignExpressionType(e, st)
	case *ast.UnaryExpression:
		return sa.getUnaryExpressionType(e, st)
	case *ast.BinaryExpression:
		return sa.getBinaryExpressionType(e, st)
	case *ast.CaseExpression:
		return sa.getCaseExpressionType(e, st)
	case *ast.IsVoidExpression:
		return "Bool"
	case *ast.MethodCallExpression:
		return sa.getMethodCallExpressionType(e, st)
	case *ast.DispatchExpression:
		return sa.getDispatchExpressionType(e, st)
	case *ast.ObjectIdentifier:
		return sa.getIdentifierType(e, st)
	case *ast.SelfExpression:
		return "SELF_TYPE"
	default:
		return "Object"
	}
}

func (sa *SemanticAnalyser) getWhileExpressionType(wexpr *ast.WhileExpression, st *SymbolTable) string {
	conditionType := sa.getExpressionType(wexpr.Condition, st)
	if conditionType != "Bool" {
		sa.errors = append(sa.errors, fmt.Sprintf("condition of if statement is of type %s, expected Bool", conditionType))
		return "Object"
	}

	return sa.getExpressionType(wexpr.Body, st)
}

func (sa *SemanticAnalyser) getBlockExpressionType(bexpr *ast.BlockExpression, st *SymbolTable) string {
	lastType := ""
	for _, expression := range bexpr.Expressions {
		lastType = sa.getExpressionType(expression, st)
	}

	return lastType
}

func (sa *SemanticAnalyser) getIfExpressionType(ifexpr *ast.IfExpression, st *SymbolTable) string {
	conditionType := sa.getExpressionType(ifexpr.Condition, st)
	if conditionType != "Bool" {
		sa.errors = append(sa.errors, fmt.Sprintf("condition of if statement is of type %s, expected Bool", conditionType))
		return "Object"
	}

	constype := sa.getExpressionType(ifexpr.Consequence, st)
	alttype := sa.getExpressionType(ifexpr.Alternative, st)

	if constype != alttype {
		sa.errors = append(sa.errors, fmt.Sprintf("ambiguous if statement return type %s vs %s", constype, alttype))
		return "Object"
	}

	return constype
}

func (sa *SemanticAnalyser) buildClassesSymboltables(program *ast.Program) {
	sa.globalSymbolTable.AddEntry("Object", &SymbolEntry{Type: "Class", Token: lexer.Token{Literal: "Object"}})
	sa.globalSymbolTable.AddEntry("Int", &SymbolEntry{Type: "Class", Token: lexer.Token{Literal: "Int"}})
	sa.globalSymbolTable.AddEntry("String", &SymbolEntry{Type: "Class", Token: lexer.Token{Literal: "String"}})
	sa.globalSymbolTable.AddEntry("Bool", &SymbolEntry{Type: "Class", Token: lexer.Token{Literal: "Bool"}})

	for _, class := range program.Classes {
		if _, ok := sa.globalSymbolTable.Lookup(class.Name.Value); ok {
			sa.errors = append(sa.errors, fmt.Sprintf("class %s is already defined", class.Name.Value))
			continue
		}

		sa.globalSymbolTable.AddEntry(class.Name.Value, &SymbolEntry{Type: "Class", Token: class.Name.Token})
	}
}

func (sa *SemanticAnalyser) buildSymboltables(program *ast.Program) {
	for _, class := range program.Classes {
		classEntry, _ := sa.globalSymbolTable.Lookup(class.Name.Value)
		classEntry.Scope = NewSymbolTable(sa.globalSymbolTable)

		for _, feature := range class.Features {
			switch f := feature.(type) {
			case *ast.Attribute:
				if _, ok := classEntry.Scope.Lookup(f.Name.Value); ok {
					sa.errors = append(sa.errors, fmt.Sprintf("attribute %s is already defined in class %s", f.Name.Value, class.Name.Value))
					continue
				}
				classEntry.Scope.AddEntry(f.Name.Value, &SymbolEntry{Token: f.Name.Token, AttrType: f.Type})
			case *ast.Method:
				methodST := NewSymbolTable(classEntry.Scope)
				classEntry.Scope.AddEntry(f.Name.Value, &SymbolEntry{Token: f.Name.Token, Scope: methodST, Method: f})
			}
		}
	}
}

// Update getNewExpressionType to handle SELF_TYPE
func (sa *SemanticAnalyser) getNewExpressionType(ne *ast.NewExpression, st *SymbolTable) string {
	if ne.Type.Value == "SELF_TYPE" {
		return "SELF_TYPE"
	}

	if _, ok := sa.globalSymbolTable.Lookup(ne.Type.Value); !ok {
		sa.errors = append(sa.errors, fmt.Sprintf("undefined type %s in new expression", ne.Type.Value))
		return "Object"
	}
	return ne.Type.Value
}

func (sa *SemanticAnalyser) getLetExpressionType(le *ast.LetExpression, st *SymbolTable) string {
	for _, b := range le.Bindings {
		sa.CheckBindingType(b, st)
	}
	return sa.getExpressionType(le.In, st)
}

func (sa *SemanticAnalyser) CheckBindingType(b *ast.LetBinding, st *SymbolTable) {
	if b.Init != nil {
		exprType := sa.getExpressionType(b.Init, st)
		if exprType != b.Type.Value {
			sa.errors = append(sa.errors, fmt.Sprintf("Let binding with wrong type %s", exprType))
		}
	}
}

func (sa *SemanticAnalyser) getAssignExpressionType(a *ast.AssignExpression, st *SymbolTable) string {
	// Get the left-hand side identifier
	var leftType string
	if identifier, ok := a.Left.(*ast.ObjectIdentifier); ok {
		// Look up the identifier in the symbol table
		if entry, ok := st.Lookup(identifier.Value); ok {
			if entry.AttrType != nil {
				leftType = entry.AttrType.Value
			} else {
				leftType = entry.Type
			}
		} else {
			sa.errors = append(sa.errors, fmt.Sprintf("undefined identifier %s in assignment", identifier.Value))
			return "Object"
		}

		// Check if we're trying to assign to 'self'
		if identifier.Value == "self" {
			sa.errors = append(sa.errors, "cannot assign to 'self'")
			return "Object"
		}
	} else {
		sa.errors = append(sa.errors, "left side of assignment must be an identifier")
		return "Object"
	}

	// Get the right-hand side type
	rightType := sa.getExpressionType(a.Right, st)

	// Check type conformance
	if !sa.isTypeConformant(rightType, leftType) {
		sa.errors = append(sa.errors, fmt.Sprintf("cannot assign expression of type %s to identifier of type %s",
			rightType, leftType))
		return "Object"
	}

	return rightType
}

func (sa *SemanticAnalyser) getUnaryExpressionType(uexpr *ast.UnaryExpression, st *SymbolTable) string {
	rightType := sa.getExpressionType(uexpr.Right, st)
	switch uexpr.Operator {
	case "~":
		if rightType != "Int" {
			sa.errors = append(sa.errors, fmt.Sprintf("bitwise negation on non-Int type: %s", rightType))
		}
		return "Int"
	case "not":
		if rightType != "Bool" {
			sa.errors = append(sa.errors, fmt.Sprintf("logical negation on non-Bool type: %s", rightType))
		}
		return "Bool"
	default:
		sa.errors = append(sa.errors, fmt.Sprintf("unknown unary operator %s", uexpr.Operator))
		return "Object"
	}
}

func isComparable(t string) bool {
	return t == "Int" || t == "Bool" || t == "String"
}

func (sa *SemanticAnalyser) getBinaryExpressionType(be *ast.BinaryExpression, st *SymbolTable) string {
	leftType := sa.getExpressionType(be.Left, st)
	rightType := sa.getExpressionType(be.Right, st) // Fix: Changed be.Left to be.Right
	switch be.Operator {
	case "+", "*", "/", "-":
		if leftType != "Int" || rightType != "Int" {
			sa.errors = append(sa.errors, fmt.Sprintf("arithmetic operation on non-Int types: %s %s %s", leftType, be.Operator, rightType))
		}
		return "Int"
	case "<", "<=", "=":
		if leftType != rightType || !isComparable(leftType) {
			sa.errors = append(sa.errors, fmt.Sprintf("comparison between incompatible types: %s %s %s", leftType, be.Operator, rightType))
		}
		return "Bool"
	default:
		sa.errors = append(sa.errors, fmt.Sprintf("unknown binary operator %s", be.Operator))
		return "Object"
	}
}

func (sa *SemanticAnalyser) getCaseExpressionType(ce *ast.CaseExpression, st *SymbolTable) string {
	types := make([]string, len(ce.Branches))
	for i, branch := range ce.Branches {
		types[i] = sa.getExpressionType(branch.Expression, st)
	}
	return sa.getLeastCommonAncestor(types)
}

// Update getMethodCallExpressionType to handle SELF_TYPE
func (sa *SemanticAnalyser) getMethodCallExpressionType(mc *ast.MethodCallExpression, st *SymbolTable) string {
	callCount["getMethodCallExpressionType"]++
	if callCount["getMethodCallExpressionType"] > 1000 {
		panic("Potential infinite recursion in getMethodCallExpressionType")
	}

	debug("Method call: %s", mc.Method.Value)
	debugDepth++
	defer func() { debugDepth-- }()

	objectType := sa.getExpressionType(mc.Object, st)
	debug("Object type: %s", objectType)
	if objectType == "SELF_TYPE" {
		objectType = sa.currentClass
	}

	// Look up the method in the class's methods
	if classEntry, ok := sa.globalSymbolTable.Lookup(objectType); ok {
		if methodEntry, ok := classEntry.Scope.Lookup(mc.Method.Value); ok {
			returnType := methodEntry.Method.Type.Value
			if returnType == "SELF_TYPE" {
				return objectType
			}
			return returnType
		}
	}

	sa.errors = append(sa.errors, fmt.Sprintf("undefined method %s for type %s", mc.Method.Value, objectType))
	return "Object"
}

func (sa *SemanticAnalyser) getDispatchExpressionType(de *ast.DispatchExpression, st *SymbolTable) string {
	// For dispatch without explicit receiver, the implicit receiver is 'self'
	classEntry, ok := sa.globalSymbolTable.Lookup(sa.currentClass)
	if !ok {
		sa.errors = append(sa.errors, fmt.Sprintf("internal error: current class %s not found", sa.currentClass))
		return "Object"
	}

	// Check method arguments
	for _, arg := range de.Arguments {
		argType := sa.getExpressionType(arg, st)
		if argType == "Object" {
			// If any argument has an error, continue checking but the call will be invalid
			continue
		}
	}

	// Look up the method in the current class's scope
	if methodEntry, ok := classEntry.Scope.Lookup(de.Method.Value); ok {
		if methodEntry.Method == nil {
			sa.errors = append(sa.errors, fmt.Sprintf("%s is not a method in class %s",
				de.Method.Value, sa.currentClass))
			return "Object"
		}

		// Check number of arguments matches method definition
		if len(de.Arguments) != len(methodEntry.Method.Formals) {
			sa.errors = append(sa.errors, fmt.Sprintf("method %s expects %d arguments but got %d",
				de.Method.Value, len(methodEntry.Method.Formals), len(de.Arguments)))
			return "Object"
		}

		// Check argument types conform to formal parameter types
		for i, arg := range de.Arguments {
			argType := sa.getExpressionType(arg, st)
			formalType := methodEntry.Method.Formals[i].Type.Value
			if !sa.isTypeConformant(argType, formalType) {
				sa.errors = append(sa.errors, fmt.Sprintf("in call to %s, argument %d type %s does not conform to parameter type %s",
					de.Method.Value, i+1, argType, formalType))
			}
		}

		// Handle SELF_TYPE in return type
		if methodEntry.Method.Type.Value == "SELF_TYPE" {
			return "SELF_TYPE"
		}
		return methodEntry.Method.Type.Value
	}

	// Check parent classes for the method
	currentClass := sa.currentClass
	for {
		parentClass, ok := sa.inheritanceGraph[currentClass]
		if !ok || parentClass == "Object" {
			break
		}
		if parentEntry, ok := sa.globalSymbolTable.Lookup(parentClass); ok {
			if methodEntry, ok := parentEntry.Scope.Lookup(de.Method.Value); ok {
				if methodEntry.Method.Type.Value == "SELF_TYPE" {
					return "SELF_TYPE"
				}
				return methodEntry.Method.Type.Value
			}
		}
		currentClass = parentClass
	}

	sa.errors = append(sa.errors, fmt.Sprintf("undefined method %s in class %s",
		de.Method.Value, sa.currentClass))
	return "Object"
}

func (sa *SemanticAnalyser) getIdentifierType(id *ast.ObjectIdentifier, st *SymbolTable) string {
	if entry, ok := st.Lookup(id.Value); ok {
		return entry.Type
	}
	sa.errors = append(sa.errors, fmt.Sprintf("undefined identifier %s", id.Value))
	return "Object"
}
