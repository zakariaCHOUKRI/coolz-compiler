package parser

import (
	"coolz-compiler/ast"
	"coolz-compiler/lexer"
	"fmt"
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
		return fmt.Sprintf("%v", e.Value)
	case *ast.ObjectIdentifier:
		return e.Value
	case *ast.BinaryExpression:
		opStr := ""
		switch e.Operator {
		case lexer.PLUS:
			opStr = "+"
		case lexer.MINUS:
			opStr = "-"
		case lexer.TIMES:
			opStr = "*"
		case lexer.DIVIDE:
			opStr = "/"
		case lexer.LT:
			opStr = "<"
		case lexer.LE:
			opStr = "<="
		case lexer.EQ:
			opStr = "="
		case lexer.ASSIGN:
			opStr = "<-"
		default:
			opStr = fmt.Sprint(e.Operator)
		}
		return fmt.Sprintf("(%s %s %s)",
			SerializeExpression(e.Left),
			opStr,
			SerializeExpression(e.Right))
	case *ast.UnaryExpression:
		opStr := ""
		switch e.Operator {
		case lexer.NEG:
			opStr = "~"
		case lexer.NOT:
			opStr = "not"
		default:
			opStr = fmt.Sprint(e.Operator)
		}
		rightExpr := SerializeExpression(e.Right)
		if rightExpr == "" {
			return ""
		}
		return fmt.Sprintf("(%s %s)", opStr, rightExpr)
	case *ast.IsVoid:
		return fmt.Sprintf("isvoid %s",
			SerializeExpression(e.Expr))
	case *ast.New:
		return fmt.Sprintf("new %s", e.Type.Value)
	case *ast.Block:
		result := "{ "
		for i, expr := range e.Expressions {
			if i > 0 {
				result += "; "
			}
			result += SerializeExpression(expr)
		}
		return result + "; }"
	case *ast.Conditional:
		return fmt.Sprintf("if %s then %s else %s fi",
			SerializeExpression(e.Condition),
			SerializeExpression(e.ThenBranch),
			SerializeExpression(e.ElseBranch))
	case *ast.Loop:
		return fmt.Sprintf("while %s loop %s pool",
			SerializeExpression(e.Condition),
			SerializeExpression(e.Body))
	case *ast.Let:
		result := "let "
		for i, decl := range e.Declarations {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("%s : %s", decl.Name.Value, decl.Type.Value)
			if decl.Init != nil {
				result += fmt.Sprintf(" <- %s", SerializeExpression(decl.Init))
			}
		}
		result += fmt.Sprintf(" in %s", SerializeExpression(e.Body))
		return result
	default:
		return fmt.Sprintf("<unknown expression type: %T>", expr)
	}
}
