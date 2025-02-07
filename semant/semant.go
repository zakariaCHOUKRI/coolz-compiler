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
	Class    *ast.Class
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

func (st *SymbolTable) GetSelfType() string {
	// Minimal placeholder; adjust if you need the actual class name
	return "SELF_TYPE"
}

type SemanticAnalyser struct {
	globalSymbolTable *SymbolTable
	errors            []string
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
		st := sa.globalSymbolTable.symbols[class.Name.Value].Scope
		sa.typeCheckClass(class, st)
	}
}

func (sa *SemanticAnalyser) typeCheckClass(cls *ast.Class, st *SymbolTable) {
	for _, feature := range cls.Features {
		switch f := feature.(type) {
		case *ast.Attribute:
			sa.typeCheckAttribute(f, st)
		case *ast.Method:
			sa.typeCheckMethod(f, st)
		}
	}
}

func (sa *SemanticAnalyser) typeCheckAttribute(attribute *ast.Attribute, st *SymbolTable) {
	if attribute.Init != nil {
		expressionType := sa.getExpressionType(attribute.Init, st)
		if expressionType != attribute.Type.Value {
			sa.errors = append(sa.errors, fmt.Sprintf("attribute %s cannot be of type %s, expected %s", attribute.Name.Value, expressionType, attribute.Type.Value))
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

		methodSt.parent.AddEntry(formal.Name.Value, &SymbolEntry{Token: formal.Name.Token, Type: formal.Type.Value})
	}

	methodBodyType := sa.getExpressionType(method.Body, methodSt)
	if methodBodyType != method.Type.Value {
		sa.errors = append(sa.errors, fmt.Sprintf("method %s is expected to return %s, found %s", method.Name.Value, method.Type.Value, methodBodyType))
	}
}

func (sa *SemanticAnalyser) isTypeConformant(type1, type2 string) bool {
	if type1 == type2 {
		return true
	}
	if type2 == "Object" {
		return true
	}
	// Walk up the parent chain to see if type1 inherits from type2
	current := type1
	for {
		entry, ok := sa.globalSymbolTable.Lookup(current)
		if !ok {
			return false
		}
		classDef := entry.Class
		if classDef == nil || classDef.Parent == nil {
			return false
		}
		parentName := classDef.Parent.Value
		if parentName == type2 {
			return true
		}
		current = parentName
	}
}

func (sa *SemanticAnalyser) getExpressionType(expression ast.Expression, st *SymbolTable) string {
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
		return sa.GetNewExpressionType(e, st)
	case *ast.LetExpression:
		return sa.GetLetExpressionType(e, st)
	case *ast.AssignExpression:
		return sa.GetAssignmentExpressionType(e, st)
	case *ast.UnaryExpression:
		return sa.GetUnaryExpressionType(e, st)
	case *ast.BinaryExpression:
		return sa.GetBinaryExpressionType(e, st)
	case *ast.CaseExpression:
		return sa.GetCaseExpressionType(e, st)
	case *ast.IsVoidExpression:
		return "Bool"
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

		sa.globalSymbolTable.AddEntry(class.Name.Value, &SymbolEntry{Type: "Class", Token: class.Name.Token, Class: class})
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

func (sa *SemanticAnalyser) GetNewExpressionType(ne *ast.NewExpression, st *SymbolTable) string {
	if ne.Type.Value == "SELF_TYPE" {
		return st.GetSelfType()
	}
	typeName := ne.Type.Value
	if _, ok := sa.globalSymbolTable.Lookup(typeName); !ok {
		sa.errors = append(sa.errors, fmt.Sprintf("undefined type %s in new expression", typeName))
		return "Object"
	}
	return typeName
}

func (sa *SemanticAnalyser) GetLetExpressionType(le *ast.LetExpression, st *SymbolTable) string {
	for _, b := range le.Bindings {
		sa.CheckBindingType(b, st)
	}
	return sa.getExpressionType(le.In, st)
}

func (sa *SemanticAnalyser) CheckBindingType(b *ast.LetBinding, st *SymbolTable) {
	exprType := sa.getExpressionType(b.Init, st)
	if exprType != b.Type.Value {
		sa.errors = append(sa.errors, fmt.Sprintf("Let binding with wrong type %s", exprType))
	}
}

func (sa *SemanticAnalyser) GetAssignmentExpressionType(a *ast.AssignExpression, st *SymbolTable) string {
	varName, ok := a.Left.(*ast.ObjectIdentifier)
	if !ok {
		sa.errors = append(sa.errors, "invalid assignment target")
		return "Object"
	}
	varEntry, found := st.Lookup(varName.Value)
	if !found {
		sa.errors = append(sa.errors, fmt.Sprintf("undefined variable %s", varName.Value))
		return "Object"
	}
	rightType := sa.getExpressionType(a.Right, st)
	if !sa.isTypeConformant(rightType, varEntry.Type) {
		sa.errors = append(sa.errors, fmt.Sprintf("type mismatch: cannot assign %s to %s", rightType, varEntry.Type))
	}
	return varEntry.Type
}

func (sa *SemanticAnalyser) GetUnaryExpressionType(uexpr *ast.UnaryExpression, st *SymbolTable) string {
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

func (sa *SemanticAnalyser) GetBinaryExpressionType(be *ast.BinaryExpression, st *SymbolTable) string {
	leftType := sa.getExpressionType(be.Left, st)
	rightType := sa.getExpressionType(be.Left, st)
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

func (sa *SemanticAnalyser) GetCaseExpressionType(ce *ast.CaseExpression, st *SymbolTable) string {
	resultType := ""
	for _, branch := range ce.Branches {
		branchType := sa.getExpressionType(branch.Expression, st)
		if resultType == "" {
			resultType = branchType
		} else {
			resultType = sa.leastCommonAncestor(resultType, branchType)
		}
	}
	if resultType == "" {
		resultType = "Object"
	}
	return resultType
}

func (sa *SemanticAnalyser) leastCommonAncestor(t1, t2 string) string {
	if sa.isTypeConformant(t1, t2) {
		return t2
	}
	if sa.isTypeConformant(t2, t1) {
		return t1
	}
	return "Object"
}
