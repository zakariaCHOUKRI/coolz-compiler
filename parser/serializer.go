package parser

import (
	"coolz-compiler/ast"
	"fmt"
	"strings"
)

func SerializeExpression(expr ast.Expression) string {
	if expr == nil {
		return ""
	}

	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return fmt.Sprintf("%d", e.Value)
	case *ast.StringLiteral:
		return fmt.Sprintf("%q", e.Value)
	case *ast.BooleanLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ast.ObjectIdentifier:
		return e.Value
	case *ast.UnaryExpression:
		return fmt.Sprintf("(%s %s)", e.Operator, SerializeExpression(e.Right))
	case *ast.BinaryExpression:
		return fmt.Sprintf("(%s %s %s)",
			SerializeExpression(e.Left),
			e.Operator,
			SerializeExpression(e.Right))
	case *ast.Conditional:
		return fmt.Sprintf("if %s then %s else %s fi",
			SerializeExpression(e.Predicate),
			SerializeExpression(e.ThenBranch),
			SerializeExpression(e.ElseBranch))
	case *ast.Loop:
		return fmt.Sprintf("while %s loop %s pool",
			SerializeExpression(e.Condition),
			SerializeExpression(e.Body))
	case *ast.Block:
		exprs := make([]string, len(e.Expressions))
		for i, expr := range e.Expressions {
			exprs[i] = SerializeExpression(expr)
		}
		return fmt.Sprintf("{ %s }", strings.Join(exprs, "; "))
	case *ast.Let:
		return fmt.Sprintf("let %s : %s <- %s in %s",
			e.VarName.Value,
			e.VarType.Value,
			SerializeExpression(e.VarInit),
			SerializeExpression(e.Body))
	case *ast.New:
		return fmt.Sprintf("new %s", e.Type.Value)
	case *ast.IsVoid:
		return fmt.Sprintf("isvoid %s", SerializeExpression(e.Expr))
	default:
		return fmt.Sprintf("<unknown:%T>", expr)
	}
}
