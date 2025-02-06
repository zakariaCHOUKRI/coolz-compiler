package visualizer

import (
	"coolz-compiler/ast"
	"encoding/json"
	"fmt"
	"strings"
)

// ToJSON converts an AST to a JSON string
func ToJSON(node ast.Node) (string, error) {
	if node == nil {
		return "", fmt.Errorf("cannot visualize nil node")
	}

	bytes, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling AST: %w", err)
	}
	return string(bytes), nil
}

// ToDOT converts an AST to GraphViz DOT format
func ToDOT(node ast.Node) string {
	var sb strings.Builder
	sb.WriteString("digraph AST {\n")
	generateDOT(&sb, node, 0)
	sb.WriteString("}\n")
	return sb.String()
}

// generateDOT implementation
func generateDOT(sb *strings.Builder, node ast.Node, id int) int {
	if node == nil {
		return id
	}

	nodeID := id
	id++

	switch n := node.(type) {
	case *ast.Program:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"Program\"];\n", nodeID))
		for _, class := range n.Classes {
			childID := generateDOT(sb, class, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d;\n", nodeID, id))
			id = childID
		}
	case *ast.Class:
		label := fmt.Sprintf("Class\\n%s", n.Name.Value)
		if n.Parent != nil {
			label += fmt.Sprintf("\\ninherits %s", n.Parent.Value)
		}
		sb.WriteString(fmt.Sprintf("  n%d [label=\"%s\"];\n", nodeID, label))
		for _, feature := range n.Features {
			childID := generateDOT(sb, feature, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d;\n", nodeID, id))
			id = childID
		}
	case *ast.Method:
		label := fmt.Sprintf("Method\\n%s\\n%s", n.Name.Value, n.Type.Value)
		sb.WriteString(fmt.Sprintf("  n%d [label=\"%s\"];\n", nodeID, label))
		for _, formal := range n.Formals {
			childID := generateDOT(sb, formal, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d;\n", nodeID, id))
			id = childID
		}
		if n.Body != nil {
			childID := generateDOT(sb, n.Body, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d;\n", nodeID, id))
			id = childID
		}
	case *ast.Formal:
		label := fmt.Sprintf("Formal\\n%s:%s", n.Name.Value, n.Type.Value)
		sb.WriteString(fmt.Sprintf("  n%d [label=\"%s\"];\n", nodeID, label))
	case *ast.Attribute:
		label := fmt.Sprintf("Attribute\\n%s:%s", n.Name.Value, n.Type.Value)
		sb.WriteString(fmt.Sprintf("  n%d [label=\"%s\"];\n", nodeID, label))
		if n.Init != nil {
			childID := generateDOT(sb, n.Init, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d;\n", nodeID, id))
			id = childID
		}
	case *ast.IntegerLiteral:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"Int\\n%d\"];\n", nodeID, n.Value))
	case *ast.StringLiteral:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"String\\n\\\"%s\\\"\"];\n", nodeID, n.Value))
	case *ast.BooleanLiteral:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"Bool\\n%t\"];\n", nodeID, n.Value))
	case *ast.BinaryExpression:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"%s\"];\n", nodeID, n.Operator))
		if n.Left != nil {
			childID := generateDOT(sb, n.Left, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d;\n", nodeID, id))
			id = childID
		}
		if n.Right != nil {
			childID := generateDOT(sb, n.Right, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d;\n", nodeID, id))
			id = childID
		}
	case *ast.UnaryExpression:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"%s\"];\n", nodeID, n.Operator))
		childID := generateDOT(sb, n.Right, id)
		sb.WriteString(fmt.Sprintf("  n%d -> n%d;\n", nodeID, id))
		id = childID
	case *ast.IfExpression:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"if\"];\n", nodeID))
		if n.Condition != nil {
			childID := generateDOT(sb, n.Condition, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d [label=\"cond\"];\n", nodeID, id))
			id = childID
		}
		if n.Consequence != nil {
			childID := generateDOT(sb, n.Consequence, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d [label=\"then\"];\n", nodeID, id))
			id = childID
		}
		if n.Alternative != nil {
			childID := generateDOT(sb, n.Alternative, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d [label=\"else\"];\n", nodeID, id))
			id = childID
		}
	case *ast.WhileExpression:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"while\"];\n", nodeID))
		if n.Condition != nil {
			childID := generateDOT(sb, n.Condition, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d [label=\"cond\"];\n", nodeID, id))
			id = childID
		}
		if n.Body != nil {
			childID := generateDOT(sb, n.Body, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d [label=\"body\"];\n", nodeID, id))
			id = childID
		}
	case *ast.BlockExpression:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"block\"];\n", nodeID))
		for _, expr := range n.Expressions {
			childID := generateDOT(sb, expr, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d;\n", nodeID, id))
			id = childID
		}
	case *ast.LetExpression:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"let\"];\n", nodeID))
		for _, binding := range n.Bindings {
			childID := generateDOT(sb, binding, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d;\n", nodeID, id))
			id = childID
		}
		if n.In != nil {
			childID := generateDOT(sb, n.In, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d [label=\"in\"];\n", nodeID, id))
			id = childID
		}
	case *ast.LetBinding:
		label := fmt.Sprintf("Binding\\n%s:%s", n.Identifier.Value, n.Type.Value)
		sb.WriteString(fmt.Sprintf("  n%d [label=\"%s\"];\n", nodeID, label))
		if n.Init != nil {
			childID := generateDOT(sb, n.Init, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d;\n", nodeID, id))
			id = childID
		}
	case *ast.CaseExpression:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"case\"];\n", nodeID))
		if n.Subject != nil {
			childID := generateDOT(sb, n.Subject, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d [label=\"subject\"];\n", nodeID, id))
			id = childID
		}
		for _, branch := range n.Branches {
			childID := generateDOT(sb, branch, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d;\n", nodeID, id))
			id = childID
		}
	case *ast.CaseBranch:
		label := fmt.Sprintf("Branch\\n%s:%s", n.Variable.Value, n.Type.Value)
		sb.WriteString(fmt.Sprintf("  n%d [label=\"%s\"];\n", nodeID, label))
		if n.Expression != nil {
			childID := generateDOT(sb, n.Expression, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d;\n", nodeID, id))
			id = childID
		}
	case *ast.MethodCallExpression:
		label := fmt.Sprintf("Call\\n%s", n.Method.Value)
		sb.WriteString(fmt.Sprintf("  n%d [label=\"%s\"];\n", nodeID, label))
		if n.Object != nil {
			childID := generateDOT(sb, n.Object, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d [label=\"obj\"];\n", nodeID, id))
			id = childID
		}
		for _, arg := range n.Arguments {
			childID := generateDOT(sb, arg, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d [label=\"arg\"];\n", nodeID, id))
			id = childID
		}
	case *ast.ObjectIdentifier:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"Id\\n%s\"];\n", nodeID, n.Value))
	case *ast.TypeIdentifier:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"Type\\n%s\"];\n", nodeID, n.Value))
	case *ast.SelfExpression:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"self\"];\n", nodeID))
	case *ast.IsVoidExpression:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"isvoid\"];\n", nodeID))
		if n.Expression != nil {
			childID := generateDOT(sb, n.Expression, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d;\n", nodeID, id))
			id = childID
		}
	case *ast.AssignExpression:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"<-\"];\n", nodeID))
		if n.Left != nil {
			childID := generateDOT(sb, n.Left, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d [label=\"left\"];\n", nodeID, id))
			id = childID
		}
		if n.Right != nil {
			childID := generateDOT(sb, n.Right, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d [label=\"right\"];\n", nodeID, id))
			id = childID
		}
	case *ast.NewExpression:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"new %s\"];\n", nodeID, n.Type.Value))
	case *ast.DispatchExpression:
		sb.WriteString(fmt.Sprintf("  n%d [label=\"dispatch\\n%s\"];\n", nodeID, n.Method.Value))
		for _, arg := range n.Arguments {
			childID := generateDOT(sb, arg, id)
			sb.WriteString(fmt.Sprintf("  n%d -> n%d [label=\"arg\"];\n", nodeID, id))
			id = childID
		}
	}

	return id
}

// Pretty prints the AST in a human-readable format
func ToPrettyString(node ast.Node) string {
	var sb strings.Builder
	prettyPrint(&sb, node, 0)
	return sb.String()
}

func prettyPrint(sb *strings.Builder, node ast.Node, indent int) {
	if node == nil {
		return
	}

	indentStr := strings.Repeat("  ", indent)

	switch n := node.(type) {
	case *ast.Program:
		sb.WriteString(indentStr + "Program\n")
		for _, class := range n.Classes {
			prettyPrint(sb, class, indent+1)
		}
	case *ast.Class:
		sb.WriteString(fmt.Sprintf("%sClass %s", indentStr, n.Name.Value))
		if n.Parent != nil {
			sb.WriteString(fmt.Sprintf(" inherits %s", n.Parent.Value))
		}
		sb.WriteString("\n")
		for _, feature := range n.Features {
			prettyPrint(sb, feature, indent+1)
		}
	case *ast.Method:
		sb.WriteString(fmt.Sprintf("%sMethod %s(", indentStr, n.Name.Value))
		for i, formal := range n.Formals {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s: %s", formal.Name.Value, formal.Type.Value))
		}
		sb.WriteString(fmt.Sprintf("): %s\n", n.Type.Value))
		prettyPrint(sb, n.Body, indent+1)
	case *ast.Attribute:
		sb.WriteString(fmt.Sprintf("%sAttribute %s: %s", indentStr, n.Name.Value, n.Type.Value))
		if n.Init != nil {
			sb.WriteString(" <- ")
			prettyPrint(sb, n.Init, 0)
		}
		sb.WriteString("\n")
	case *ast.IntegerLiteral:
		sb.WriteString(fmt.Sprintf("%d", n.Value))
	case *ast.StringLiteral:
		sb.WriteString(fmt.Sprintf("%q", n.Value))
	case *ast.BooleanLiteral:
		sb.WriteString(fmt.Sprintf("%t", n.Value))
	case *ast.BinaryExpression:
		sb.WriteString("(")
		prettyPrint(sb, n.Left, 0)
		sb.WriteString(fmt.Sprintf(" %s ", n.Operator))
		prettyPrint(sb, n.Right, 0)
		sb.WriteString(")")
	case *ast.IfExpression:
		sb.WriteString(fmt.Sprintf("%sif ", indentStr))
		prettyPrint(sb, n.Condition, 0)
		sb.WriteString(" then\n")
		prettyPrint(sb, n.Consequence, indent+1)
		sb.WriteString(fmt.Sprintf("%selse\n", indentStr))
		prettyPrint(sb, n.Alternative, indent+1)
		sb.WriteString(fmt.Sprintf("%sfi\n", indentStr))
	case *ast.WhileExpression:
		sb.WriteString(fmt.Sprintf("%swhile ", indentStr))
		prettyPrint(sb, n.Condition, 0)
		sb.WriteString(" loop\n")
		prettyPrint(sb, n.Body, indent+1)
		sb.WriteString(fmt.Sprintf("%spool\n", indentStr))
	case *ast.LetExpression:
		sb.WriteString(fmt.Sprintf("%slet ", indentStr))
		for i, binding := range n.Bindings {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s: %s", binding.Identifier.Value, binding.Type.Value))
			if binding.Init != nil {
				sb.WriteString(" <- ")
				prettyPrint(sb, binding.Init, 0)
			}
		}
		sb.WriteString(" in\n")
		prettyPrint(sb, n.In, indent+1)
	case *ast.CaseExpression:
		sb.WriteString(fmt.Sprintf("%scase ", indentStr))
		prettyPrint(sb, n.Subject, 0)
		sb.WriteString(" of\n")
		for _, branch := range n.Branches {
			sb.WriteString(fmt.Sprintf("%s  %s: %s => ", indentStr,
				branch.Variable.Value, branch.Type.Value))
			prettyPrint(sb, branch.Expression, 0)
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("%sesac\n", indentStr))
	case *ast.MethodCallExpression:
		prettyPrint(sb, n.Object, 0)
		sb.WriteString(fmt.Sprintf(".%s(", n.Method.Value))
		for i, arg := range n.Arguments {
			if i > 0 {
				sb.WriteString(", ")
			}
			prettyPrint(sb, arg, 0)
		}
		sb.WriteString(")")
	case *ast.DispatchExpression:
		sb.WriteString(fmt.Sprintf("%s(", n.Method.Value))
		for i, arg := range n.Arguments {
			if i > 0 {
				sb.WriteString(", ")
			}
			prettyPrint(sb, arg, 0)
		}
		sb.WriteString(")")
	case *ast.IsVoidExpression:
		sb.WriteString("isvoid ")
		prettyPrint(sb, n.Expression, 0)
	case *ast.NewExpression:
		sb.WriteString(fmt.Sprintf("new %s", n.Type.Value))
	case *ast.AssignExpression:
		prettyPrint(sb, n.Left, 0)
		sb.WriteString(" <- ")
		prettyPrint(sb, n.Right, 0)
	case *ast.UnaryExpression:
		sb.WriteString(n.Operator)
		sb.WriteString(" ")
		prettyPrint(sb, n.Right, 0)
	case *ast.BlockExpression:
		sb.WriteString("{\n")
		for _, expr := range n.Expressions {
			sb.WriteString(indentStr + "  ")
			prettyPrint(sb, expr, indent+1)
			sb.WriteString(";\n")
		}
		sb.WriteString(indentStr + "}")
	}
}

// Add a new helper function for escaping DOT labels
func escapeDOTLabel(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\"", "\\\""), "\n", "\\n")
}
