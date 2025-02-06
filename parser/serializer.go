package parser

import (
	"coolz-compiler/ast"
	"fmt"
	"strings"
)

func SerializeExpression(exp ast.Expression) string {
	if exp == nil {
		return "<nil>"
	}

	switch e := exp.(type) {
	case *ast.IntegerLiteral:
		return e.Token.Literal
	case *ast.StringLiteral:
		return fmt.Sprintf("%q", e.Token.Literal) // Add quotes around string literals
	case *ast.BooleanLiteral:
		return e.Token.Literal
	case *ast.ObjectIdentifier:
		return e.Value
	case *ast.UnaryExpression:
		return fmt.Sprintf("(%s %s)", e.Operator, SerializeExpression(e.Right))
	case *ast.BinaryExpression:
		return fmt.Sprintf("(%s %s %s)",
			SerializeExpression(e.Left),
			e.Operator,
			SerializeExpression(e.Right))
	case *ast.IsVoidExpression:
		return fmt.Sprintf("isvoid %s", SerializeExpression(e.Expression))
	case *ast.NewExpression:
		return fmt.Sprintf("new %s", e.Type.Value)
	case *ast.AssignExpression:
		return fmt.Sprintf("(%s <- %s)",
			SerializeExpression(e.Left),
			SerializeExpression(e.Right))
	case *ast.IfExpression:
		return fmt.Sprintf("if %s then %s else %s fi",
			SerializeExpression(e.Condition),
			SerializeExpression(e.Consequence),
			SerializeExpression(e.Alternative))
	case *ast.WhileExpression:
		return fmt.Sprintf("while %s loop %s pool",
			SerializeExpression(e.Condition),
			SerializeExpression(e.Body))
	case *ast.MethodCallExpression:
		args := make([]string, len(e.Arguments))
		for i, arg := range e.Arguments {
			args[i] = SerializeExpression(arg)
		}
		return fmt.Sprintf("((%s . %s))",
			SerializeExpression(e.Object),
			e.Method.Value)
	case *ast.SelfExpression:
		return "self"
	case *ast.DispatchExpression:
		args := make([]string, len(e.Arguments))
		for i, arg := range e.Arguments {
			args[i] = SerializeExpression(arg)
		}
		if len(args) == 0 {
			return fmt.Sprintf("%s()", e.Method.Value)
		}
		return fmt.Sprintf("%s(%s)", e.Method.Value, strings.Join(args, ", "))
	case *ast.LetExpression:
		bindings := make([]string, len(e.Bindings))
		for i, binding := range e.Bindings {
			if binding.Init != nil {
				bindings[i] = fmt.Sprintf("%s:%s <- %s",
					binding.Identifier.Value,
					binding.Type.Value,
					SerializeExpression(binding.Init))
			} else {
				bindings[i] = fmt.Sprintf("%s:%s",
					binding.Identifier.Value,
					binding.Type.Value)
			}
		}
		return fmt.Sprintf("let %s in %s",
			strings.Join(bindings, ", "),
			SerializeExpression(e.In))
	case *ast.BlockExpression:
		exprs := make([]string, len(e.Expressions))
		for i, expr := range e.Expressions {
			exprs[i] = SerializeExpression(expr)
		}
		return "{ " + strings.Join(exprs, "; ") + " }"
	case *ast.CaseExpression:
		branches := make([]string, len(e.Branches))
		for i, branch := range e.Branches {
			branches[i] = fmt.Sprintf("%s : %s => %s",
				branch.Variable.Value,
				branch.Type.Value,
				SerializeExpression(branch.Expression))
		}
		return fmt.Sprintf("case %s of %s esac",
			SerializeExpression(e.Subject),
			strings.Join(branches, "; "))
	default:
		return fmt.Sprintf("%T", exp)
	}
}
