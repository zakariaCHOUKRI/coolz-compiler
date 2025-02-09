package parser

import (
	"coolz-compiler/ast"
	"coolz-compiler/lexer"
	"fmt"
	"strings"
)

func getOperatorSymbol(t lexer.TokenType) string {
	switch t {
	case lexer.PLUS:
		return "+"
	case lexer.MINUS:
		return "-"
	case lexer.TIMES:
		return "*"
	case lexer.DIVIDE:
		return "/"
	case lexer.LT:
		return "<"
	case lexer.LE:
		return "<="
	case lexer.EQ:
		return "="
	case lexer.NOT:
		return "not"
	case lexer.NEG:
		return "~"
	default:
		return t.String()
	}
}

func SerializeExpression(node ast.Expression) string {
	switch exp := node.(type) {
	case *ast.IntegerLiteral:
		return fmt.Sprintf("%d", exp.Value)
	case *ast.StringLiteral:
		return fmt.Sprintf("%q", exp.Value)
	case *ast.BooleanLiteral:
		return fmt.Sprintf("%t", exp.Value)
	case *ast.ObjectIdentifier:
		return exp.Value
	case *ast.UnaryExpression:
		return fmt.Sprintf("(%s %s)", getOperatorSymbol(exp.Operator), SerializeExpression(exp.Right))
	case *ast.BinaryExpression:
		return fmt.Sprintf("(%s %s %s)", SerializeExpression(exp.Left), getOperatorSymbol(exp.Operator), SerializeExpression(exp.Right))
	case *ast.Assignment:
		return fmt.Sprintf("(%s <- %s)", SerializeExpression(exp.Left), SerializeExpression(exp.Value))
	case *ast.Conditional:
		return fmt.Sprintf("if %s then %s else %s fi", SerializeExpression(exp.Predicate), SerializeExpression(exp.ThenBranch), SerializeExpression(exp.ElseBranch))
	case *ast.Loop:
		return fmt.Sprintf("while %s loop %s pool", SerializeExpression(exp.Condition), SerializeExpression(exp.Body))
	case *ast.Block:
		expressions := []string{}
		for _, expr := range exp.Expressions {
			expressions = append(expressions, SerializeExpression(expr))
		}
		return fmt.Sprintf("{ %s }", strings.Join(expressions, "; "))
	case *ast.Let:
		return fmt.Sprintf("let %s : %s <- %s in %s", SerializeExpression(exp.VarName), exp.VarType.Value, SerializeExpression(exp.VarInit), SerializeExpression(exp.Body))
	case *ast.New:
		return fmt.Sprintf("new %s", exp.Type.Value)
	case *ast.IsVoid:
		return fmt.Sprintf("isvoid %s", SerializeExpression(exp.Expr))
	default:
		return fmt.Sprintf("<unknown:%T>", exp)
	}
}
