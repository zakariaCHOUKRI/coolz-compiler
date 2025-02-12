package semant

import (
	"coolz-compiler/ast"
	"coolz-compiler/lexer"
	"fmt"
)

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
	Parent   string // Track parent class name
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
	currentClass      string // Track current class during type checking
}

func NewSemanticAnalyser() *SemanticAnalyser {
	return &SemanticAnalyser{
		globalSymbolTable: NewSymbolTable(nil),
		errors:            []string{},
	}
}

func (sa *SemanticAnalyser) Errors() []string {
	return sa.errors
}

func (sa *SemanticAnalyser) Analyze(program *ast.Program) {
	sa.buildClassesSymboltables(program)
	sa.buildSymboltables(program)
	sa.typeCheck(program)
}

func (sa *SemanticAnalyser) typeCheck(program *ast.Program) {
	for _, class := range program.Classes {
		classEntry, _ := sa.globalSymbolTable.Lookup(class.Name.Value)
		sa.typeCheckClass(class, classEntry.Scope)
	}
}

func (sa *SemanticAnalyser) typeCheckClass(cls *ast.Class, st *SymbolTable) {
	sa.currentClass = cls.Name.Value
	defer func() { sa.currentClass = "" }()

	for _, feature := range cls.Features {
		switch f := feature.(type) {
		case *ast.Attribute:
			sa.typeCheckAttribute(f, st)
		case *ast.Method:
			sa.typeCheckMethod(f, st)
		}
	}
}

func (sa *SemanticAnalyser) typeCheckAttribute(attr *ast.Attribute, st *SymbolTable) {
	if attr.Init != nil {
		exprType := sa.getExpressionType(attr.Init, st)
		expectedType := attr.Type.Value
		if expectedType == "SELF_TYPE" {
			expectedType = sa.currentClass
		}
		if !sa.isTypeConformant(exprType, expectedType) {
			sa.errors = append(sa.errors, fmt.Sprintf("attribute %s cannot be of type %s, expected %s",
				attr.Name.Value, exprType, expectedType))
		}
	}
}

func (sa *SemanticAnalyser) typeCheckMethod(method *ast.Method, st *SymbolTable) {
	methodEntry, _ := st.Lookup(method.Name.Value)
	methodSt := methodEntry.Scope

	// Check return type conformance
	exprType := sa.getExpressionType(method.Body, methodSt)
	expectedType := method.Type.Value
	if expectedType == "SELF_TYPE" {
		expectedType = sa.currentClass
	}
	if !sa.isTypeConformant(exprType, expectedType) {
		sa.errors = append(sa.errors, fmt.Sprintf("method %s expects return type %s, got %s",
			method.Name.Value, expectedType, exprType))
	}
}

func (sa *SemanticAnalyser) isTypeConformant(subType, superType string) bool {
	if subType == superType {
		return true
	}
	if subType == "SELF_TYPE" {
		return sa.currentClass != "" && sa.isTypeConformant(sa.currentClass, superType)
	}

	entry, ok := sa.globalSymbolTable.Lookup(subType)
	if !ok {
		return false
	}

	current := entry.Parent
	for current != "" {
		if current == superType {
			return true
		}
		entry, ok = sa.globalSymbolTable.Lookup(current)
		if !ok {
			break
		}
		current = entry.Parent
	}

	return false
}

func (sa *SemanticAnalyser) getExpressionType(expr ast.Expression, st *SymbolTable) string {
	switch e := expr.(type) {
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
		return sa.GetNewExpressionType(e, st)
	case *ast.LetExpression:
		return sa.GetLetExpressionType(e, st)
	case *ast.Assignment:
		return sa.GetAssignmentExpressionType(e, st)
	case *ast.UnaryExpression:
		return sa.GetUnaryExpressionType(e, st)
	case *ast.BinaryExpression:
		return sa.GetBinaryExpressionType(e, st)
	case *ast.CaseExpression:
		return sa.GetCaseExpressionType(e, st)
	case *ast.IsVoidExpression:
		return "Bool"
	case *ast.ObjectIdentifier:
		return sa.getObjectIdentifierType(e, st)
	case *ast.Self:
		return "SELF_TYPE"
	default:
		return "Object"
	}
}

func (sa *SemanticAnalyser) getObjectIdentifierType(oi *ast.ObjectIdentifier, st *SymbolTable) string {
	entry, ok := st.Lookup(oi.Value)
	if !ok {
		sa.errors = append(sa.errors, fmt.Sprintf("undefined identifier %s", oi.Value))
		return "Object"
	}
	return entry.Type
}

func (sa *SemanticAnalyser) buildClassesSymboltables(program *ast.Program) {
	// Predefine basic classes
	sa.globalSymbolTable.AddEntry("Object", &SymbolEntry{Type: "Class", Parent: ""})
	sa.globalSymbolTable.AddEntry("Int", &SymbolEntry{Type: "Class", Parent: "Object"})
	sa.globalSymbolTable.AddEntry("String", &SymbolEntry{Type: "Class", Parent: "Object"})
	sa.globalSymbolTable.AddEntry("Bool", &SymbolEntry{Type: "Class", Parent: "Object"})

	for _, class := range program.Classes {
		if _, ok := sa.globalSymbolTable.Lookup(class.Name.Value); ok {
			sa.errors = append(sa.errors, fmt.Sprintf("class %s redefined", class.Name.Value))
			continue
		}

		parent := "Object"
		if class.Parent != nil {
			parent = class.Parent.Value
		} else if class.Name.Value == "Object" {
			parent = ""
		}

		sa.globalSymbolTable.AddEntry(class.Name.Value, &SymbolEntry{
			Type:   "Class",
			Token:  class.Name.Token,
			Parent: parent,
		})
	}
}

func (sa *SemanticAnalyser) buildSymboltables(program *ast.Program) {
	for _, class := range program.Classes {
		classEntry, _ := sa.globalSymbolTable.Lookup(class.Name.Value)
		classEntry.Scope = NewSymbolTable(sa.globalSymbolTable)

		// Add attributes and methods
		for _, feature := range class.Features {
			switch f := feature.(type) {
			case *ast.Attribute:
				if _, ok := classEntry.Scope.Lookup(f.Name.Value); ok {
					sa.errors = append(sa.errors, fmt.Sprintf("attribute %s redefined", f.Name.Value))
					continue
				}
				classEntry.Scope.AddEntry(f.Name.Value, &SymbolEntry{
					Type:     f.Type.Value,
					Token:    f.Name.Token,
					AttrType: f.Type,
				})

			case *ast.Method:
				methodSt := NewSymbolTable(classEntry.Scope)
				// Add formals to method's scope
				for _, formal := range f.Formals {
					methodSt.AddEntry(formal.Name.Value, &SymbolEntry{
						Type:  formal.Type.Value,
						Token: formal.Name.Token,
					})
				}
				classEntry.Scope.AddEntry(f.Name.Value, &SymbolEntry{
					Token:  f.Name.Token,
					Method: f,
					Scope:  methodSt,
					Type:   f.Type.Value,
				})
			}
		}
	}
}

func (sa *SemanticAnalyser) GetNewExpressionType(ne *ast.NewExpression, st *SymbolTable) string {
	if ne.Type.Value == "SELF_TYPE" {
		if sa.currentClass == "" {
			sa.errors = append(sa.errors, "SELF_TYPE used outside class")
			return "Object"
		}
		return sa.currentClass
	}

	if _, ok := sa.globalSymbolTable.Lookup(ne.Type.Value); !ok {
		sa.errors = append(sa.errors, fmt.Sprintf("undefined type %s", ne.Type.Value))
		return "Object"
	}
	return ne.Type.Value
}

func (sa *SemanticAnalyser) GetAssignmentExpressionType(a *ast.Assignment, st *SymbolTable) string {
	left, ok := a.Left.(*ast.ObjectIdentifier)
	if !ok {
		sa.errors = append(sa.errors, "assignment to non-identifier")
		return "Object"
	}

	entry, exists := st.Lookup(left.Value)
	if !exists {
		sa.errors = append(sa.errors, fmt.Sprintf("undefined variable %s", left.Value))
		return "Object"
	}

	valueType := sa.getExpressionType(a.Value, st)
	if !sa.isTypeConformant(valueType, entry.Type) {
		sa.errors = append(sa.errors, fmt.Sprintf("type %s does not conform to %s", valueType, entry.Type))
	}
	return valueType
}

func (sa *SemanticAnalyser) GetCaseExpressionType(ce *ast.CaseExpression, st *SymbolTable) string {
	var branchTypes []string
	for _, branch := range ce.Branches {
		// Check branch type validity
		if _, ok := sa.globalSymbolTable.Lookup(branch.Type.Value); !ok {
			sa.errors = append(sa.errors, fmt.Sprintf("undefined type %s", branch.Type.Value))
			continue
		}

		// Create branch scope
		branchSt := NewSymbolTable(st)
		branchSt.AddEntry(branch.Identifier.Value, &SymbolEntry{
			Type:  branch.Type.Value,
			Token: branch.Identifier.Token,
		})

		// Get branch expression type
		exprType := sa.getExpressionType(branch.Expr, branchSt)
		branchTypes = append(branchTypes, exprType)
	}

	if len(branchTypes) == 0 {
		return "Object"
	}
	return sa.joinTypes(branchTypes)
}

func (sa *SemanticAnalyser) joinTypes(types []string) string {
	if len(types) == 0 {
		return "Object"
	}

	join := types[0]
	for _, t := range types[1:] {
		join = sa.findCommonAncestor(join, t)
	}
	return join
}

func (sa *SemanticAnalyser) findCommonAncestor(a, b string) string {
	ancestorsA := sa.getAncestors(a)
	ancestorsB := sa.getAncestors(b)

	// Find the first common ancestor in A's ancestor list
	for _, ancestorA := range ancestorsA {
		for _, ancestorB := range ancestorsB {
			if ancestorA == ancestorB {
				return ancestorA
			}
		}
	}
	return "Object" // Fallback
}

func (sa *SemanticAnalyser) getAncestors(typ string) []string {
	var ancestors []string
	current := typ
	for {
		ancestors = append(ancestors, current)
		entry, ok := sa.globalSymbolTable.Lookup(current)
		if !ok || entry.Parent == "" {
			break
		}
		current = entry.Parent
	}
	return ancestors
}

// ... (Other existing functions like GetLetExpressionType, GetUnaryExpressionType, etc. remain with similar updates)

func (sa *SemanticAnalyser) GetLetExpressionType(le *ast.LetExpression, st *SymbolTable) string {
	for _, binding := range le.Bindings {
		// Check initialization expression type
		if binding.Init != nil {
			initType := sa.getExpressionType(binding.Init, st)
			if !sa.isTypeConformant(initType, binding.Type.Value) {
				sa.errors = append(sa.errors, fmt.Sprintf("let binding %s: type %s does not conform to %s",
					binding.Identifier.Value, initType, binding.Type.Value))
			}
		}

		// Add the binding to the scope
		st.AddEntry(binding.Identifier.Value, &SymbolEntry{
			Type:  binding.Type.Value,
			Token: binding.Identifier.Token,
		})
	}

	// Return the type of the 'in' expression
	return sa.getExpressionType(le.In, st)
}

func (sa *SemanticAnalyser) GetUnaryExpressionType(ue *ast.UnaryExpression, st *SymbolTable) string {
	rightType := sa.getExpressionType(ue.Right, st)

	switch ue.Operator {
	case "~":
		if rightType != "Int" {
			sa.errors = append(sa.errors, fmt.Sprintf("bitwise negation (~) requires Int, got %s", rightType))
		}
		return "Int"
	case "not":
		if rightType != "Bool" {
			sa.errors = append(sa.errors, fmt.Sprintf("logical negation (not) requires Bool, got %s", rightType))
		}
		return "Bool"
	default:
		sa.errors = append(sa.errors, fmt.Sprintf("unknown unary operator %s", ue.Operator))
		return "Object"
	}
}

func (sa *SemanticAnalyser) GetBinaryExpressionType(be *ast.BinaryExpression, st *SymbolTable) string {
	leftType := sa.getExpressionType(be.Left, st)
	rightType := sa.getExpressionType(be.Right, st)

	switch be.Operator {
	case "+", "-", "*", "/":
		if leftType != "Int" || rightType != "Int" {
			sa.errors = append(sa.errors, fmt.Sprintf("arithmetic operator %s requires Int, got %s and %s",
				be.Operator, leftType, rightType))
		}
		return "Int"
	case "<", "<=":
		if leftType != "Int" || rightType != "Int" {
			sa.errors = append(sa.errors, fmt.Sprintf("comparison operator %s requires Int, got %s and %s",
				be.Operator, leftType, rightType))
		}
		return "Bool"
	case "=":
		if !sa.isTypeConformant(leftType, rightType) && !sa.isTypeConformant(rightType, leftType) {
			sa.errors = append(sa.errors, fmt.Sprintf("equality operator = requires conforming types, got %s and %s",
				leftType, rightType))
		}
		return "Bool"
	default:
		sa.errors = append(sa.errors, fmt.Sprintf("unknown binary operator %s", be.Operator))
		return "Object"
	}
}

func (sa *SemanticAnalyser) getBlockExpressionType(be *ast.BlockExpression, st *SymbolTable) string {
	var lastType string
	for _, expr := range be.Expressions {
		lastType = sa.getExpressionType(expr, st)
	}
	return lastType
}

func (sa *SemanticAnalyser) getIfExpressionType(ie *ast.IfExpression, st *SymbolTable) string {
	condType := sa.getExpressionType(ie.Condition, st)
	if condType != "Bool" {
		sa.errors = append(sa.errors, fmt.Sprintf("if condition must be Bool, got %s", condType))
	}

	thenType := sa.getExpressionType(ie.Consequence, st)
	elseType := sa.getExpressionType(ie.Alternative, st)

	return sa.findCommonAncestor(thenType, elseType)
}

func (sa *SemanticAnalyser) getWhileExpressionType(we *ast.WhileExpression, st *SymbolTable) string {
	condType := sa.getExpressionType(we.Condition, st)
	if condType != "Bool" {
		sa.errors = append(sa.errors, fmt.Sprintf("while condition must be Bool, got %s", condType))
	}

	// While expressions always return Object (void)
	return "Object"
}
