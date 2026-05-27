package printer

import (
	astpkg "compilerlabs/ast"
	"fmt"
	"strings"
)

func PrintProgram(statements []astpkg.Statement) string {
	var b strings.Builder
	for i, stmt := range statements {
		writeStatement(&b, stmt, 0)
		if i < len(statements)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}
func writeStatement(b *strings.Builder, stmt astpkg.Statement, indent int) {
	p := strings.Repeat("  ", indent)
	switch s := stmt.(type) {
	case *astpkg.VarStatement:
		typeInfo := ""
		if s.TypeAnnotation != nil {
			typeInfo = fmt.Sprintf(", type=%s", s.TypeAnnotation.String())
		}
		fmt.Fprintf(b, "%sVarStatement(name=%s%s, span=%s)\n", p, s.Name, typeInfo, s.Span())
		if s.Initializer != nil {
			fmt.Fprintf(b, "%s  Initializer:\n", p)
			writeExpression(b, s.Initializer, indent+2)
		}
	case *astpkg.PrintStatement:
		fmt.Fprintf(b, "%sPrintStatement(span=%s)\n", p, s.Span())
		writeExpression(b, s.Expr, indent+1)
	case *astpkg.ExpressionStatement:
		fmt.Fprintf(b, "%sExpressionStatement(span=%s)\n", p, s.Span())
		writeExpression(b, s.Expr, indent+1)
	case *astpkg.BlockStatement:
		fmt.Fprintf(b, "%sBlockStatement(span=%s)\n", p, s.Span())
		for _, inner := range s.Statements {
			writeStatement(b, inner, indent+1)
		}
	case *astpkg.IfStatement:
		fmt.Fprintf(b, "%sIfStatement(span=%s)\n", p, s.Span())
		fmt.Fprintf(b, "%s  Condition:\n", p)
		writeExpression(b, s.Condition, indent+2)
		fmt.Fprintf(b, "%s  Then:\n", p)
		writeStatement(b, s.ThenBranch, indent+2)
		if s.ElseBranch != nil {
			fmt.Fprintf(b, "%s  Else:\n", p)
			writeStatement(b, s.ElseBranch, indent+2)
		}
	case *astpkg.WhileStatement:
		fmt.Fprintf(b, "%sWhileStatement(span=%s)\n", p, s.Span())
		fmt.Fprintf(b, "%s  Condition:\n", p)
		writeExpression(b, s.Condition, indent+2)
		fmt.Fprintf(b, "%s  Body:\n", p)
		writeStatement(b, s.Body, indent+2)
	case *astpkg.FunctionStatement:
		params := make([]string, len(s.Parameters))
		for idx, param := range s.Parameters {
			params[idx] = param
			if idx < len(s.ParameterTypes) && s.ParameterTypes[idx] != nil {
				params[idx] = fmt.Sprintf("%s:%s", param, s.ParameterTypes[idx].String())
			}
		}
		returnInfo := ""
		if s.ReturnType != nil {
			returnInfo = fmt.Sprintf(", return=%s", s.ReturnType.String())
		}
		fmt.Fprintf(b, "%sFunctionStatement(name=%s, params=%v%s, span=%s)\n", p, s.Name, params, returnInfo, s.Span())
		writeStatement(b, s.Body, indent+1)
	case *astpkg.ReturnStatement:
		fmt.Fprintf(b, "%sReturnStatement(span=%s)\n", p, s.Span())
		if s.Value != nil {
			writeExpression(b, s.Value, indent+1)
		}
	default:
		fmt.Fprintf(b, "%s<unknown statement %T>\n", p, s)
	}
}
func writeExpression(b *strings.Builder, expr astpkg.Expression, indent int) {
	p := strings.Repeat("  ", indent)
	switch e := expr.(type) {
	case *astpkg.NumberExpression:
		fmt.Fprintf(b, "%sNumberExpression(value=%v, span=%s)\n", p, e.Value, e.Span())
	case *astpkg.StringExpression:
		fmt.Fprintf(b, "%sStringExpression(value=%q, span=%s)\n", p, e.Value, e.Span())
	case *astpkg.BoolExpression:
		fmt.Fprintf(b, "%sBoolExpression(value=%v, span=%s)\n", p, e.Value, e.Span())
	case *astpkg.VariableExpression:
		fmt.Fprintf(b, "%sVariableExpression(name=%s, span=%s)\n", p, e.Name, e.Span())
	case *astpkg.GroupingExpression:
		fmt.Fprintf(b, "%sGroupingExpression(span=%s)\n", p, e.Span())
		writeExpression(b, e.Expr, indent+1)
	case *astpkg.ArrayExpression:
		fmt.Fprintf(b, "%sArrayExpression(span=%s)\n", p, e.Span())
		for _, el := range e.Elements {
			writeExpression(b, el, indent+1)
		}
	case *astpkg.IndexExpression:
		fmt.Fprintf(b, "%sIndexExpression(span=%s)\n", p, e.Span())
		fmt.Fprintf(b, "%s  Array:\n", p)
		writeExpression(b, e.Array, indent+2)
		fmt.Fprintf(b, "%s  Index:\n", p)
		writeExpression(b, e.Index, indent+2)
	case *astpkg.IndexAssignExpression:
		fmt.Fprintf(b, "%sIndexAssignExpression(span=%s)\n", p, e.Span())
		fmt.Fprintf(b, "%s  Array:\n", p)
		writeExpression(b, e.Array, indent+2)
		fmt.Fprintf(b, "%s  Index:\n", p)
		writeExpression(b, e.Index, indent+2)
		fmt.Fprintf(b, "%s  Value:\n", p)
		writeExpression(b, e.Value, indent+2)
	case *astpkg.UnaryExpression:
		fmt.Fprintf(b, "%sUnaryExpression(op=%s, span=%s)\n", p, e.Operator.Lexeme, e.Span())
		writeExpression(b, e.Right, indent+1)
	case *astpkg.BinaryExpression:
		fmt.Fprintf(b, "%sBinaryExpression(op=%s, span=%s)\n", p, e.Operator.Lexeme, e.Span())
		fmt.Fprintf(b, "%s  Left:\n", p)
		writeExpression(b, e.Left, indent+2)
		fmt.Fprintf(b, "%s  Right:\n", p)
		writeExpression(b, e.Right, indent+2)
	case *astpkg.AssignExpression:
		fmt.Fprintf(b, "%sAssignExpression(name=%s, span=%s)\n", p, e.Name, e.Span())
		writeExpression(b, e.Value, indent+1)
	case *astpkg.CallExpression:
		fmt.Fprintf(b, "%sCallExpression(callee=%s, span=%s)\n", p, e.CalleeName, e.Span())
		for _, arg := range e.Arguments {
			writeExpression(b, arg, indent+1)
		}
	default:
		fmt.Fprintf(b, "%s<unknown expression %T>\n", p, e)
	}
}
