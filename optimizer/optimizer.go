package optimizer

import (
	"fmt"
	"math"

	astpkg "compilerlabs/ast"
	"compilerlabs/source"
	"compilerlabs/token"
)

type Optimizer struct{}

func New() *Optimizer { return &Optimizer{} }
func (o *Optimizer) Optimize(statements []astpkg.Statement) []astpkg.Statement {
	return o.optimizeStatements(statements)
}

func (o *Optimizer) optimizeStatements(statements []astpkg.Statement) []astpkg.Statement {
	out := make([]astpkg.Statement, 0, len(statements))
	for _, stmt := range statements {
		optimized := o.optimizeStatement(stmt)
		if optimized == nil {
			continue
		}
		out = append(out, optimized)
		if _, ok := optimized.(*astpkg.ReturnStatement); ok {
			break
		}
	}
	return out
}

func (o *Optimizer) optimizeStatement(stmt astpkg.Statement) astpkg.Statement {
	switch s := stmt.(type) {
	case *astpkg.VarStatement:
		if s.Initializer != nil {
			s.Initializer = o.foldExpression(s.Initializer)
		}
		return s
	case *astpkg.PrintStatement:
		s.Expr = o.foldExpression(s.Expr)
		return s
	case *astpkg.ExpressionStatement:
		s.Expr = o.foldExpression(s.Expr)
		if isPure(s.Expr) {
			return nil
		}
		return s
	case *astpkg.BlockStatement:
		s.Statements = o.optimizeStatements(s.Statements)
		return s
	case *astpkg.IfStatement:
		s.Condition = o.foldExpression(s.Condition)
		if cond, ok := boolLiteral(s.Condition); ok {
			if cond {
				return o.optimizeStatement(s.ThenBranch)
			}
			if s.ElseBranch != nil {
				return o.optimizeStatement(s.ElseBranch)
			}
			return nil
		}
		s.ThenBranch = o.optimizeStatement(s.ThenBranch)
		if s.ElseBranch != nil {
			s.ElseBranch = o.optimizeStatement(s.ElseBranch)
		}
		return s
	case *astpkg.WhileStatement:
		s.Condition = o.foldExpression(s.Condition)
		if cond, ok := boolLiteral(s.Condition); ok && !cond {
			return nil
		}
		s.Body = o.optimizeStatement(s.Body)
		return s
	case *astpkg.FunctionStatement:
		s.Body.Statements = o.optimizeStatements(s.Body.Statements)
		return s
	case *astpkg.ReturnStatement:
		if s.Value != nil {
			s.Value = o.foldExpression(s.Value)
		}
		return s
	default:
		return stmt
	}
}

func (o *Optimizer) foldExpression(expr astpkg.Expression) astpkg.Expression {
	switch e := expr.(type) {
	case *astpkg.GroupingExpression:
		e.Expr = o.foldExpression(e.Expr)
		return e.Expr
	case *astpkg.UnaryExpression:
		e.Right = o.foldExpression(e.Right)
		if v, ok := literalValue(e.Right); ok {
			switch e.Operator.Type {
			case token.MINUS:
				if n, ok := v.(float64); ok {
					return &astpkg.NumberExpression{BaseNode: astpkg.BaseNode{NodeSpan: e.Span()}, Value: -n}
				}
			case token.BANG:
				if b, ok := v.(bool); ok {
					return &astpkg.BoolExpression{BaseNode: astpkg.BaseNode{NodeSpan: e.Span()}, Value: !b}
				}
			}
		}
		return e
	case *astpkg.BinaryExpression:
		e.Left = o.foldExpression(e.Left)
		e.Right = o.foldExpression(e.Right)
		left, lok := literalValue(e.Left)
		right, rok := literalValue(e.Right)
		if lok && rok {
			if folded := foldBinary(e.Operator.Type, e.Operator.Lexeme, left, right, e.Span()); folded != nil {
				return folded
			}
		}
		return e
	case *astpkg.ArrayExpression:
		for idx, element := range e.Elements {
			e.Elements[idx] = o.foldExpression(element)
		}
		return e
	case *astpkg.IndexExpression:
		e.Array = o.foldExpression(e.Array)
		e.Index = o.foldExpression(e.Index)
		return e
	case *astpkg.IndexAssignExpression:
		e.Array = o.foldExpression(e.Array)
		e.Index = o.foldExpression(e.Index)
		e.Value = o.foldExpression(e.Value)
		return e
	case *astpkg.AssignExpression:
		e.Value = o.foldExpression(e.Value)
		return e
	case *astpkg.CallExpression:
		for idx, arg := range e.Arguments {
			e.Arguments[idx] = o.foldExpression(arg)
		}
		return e
	default:
		return expr
	}
}

func literalValue(expr astpkg.Expression) (any, bool) {
	switch e := expr.(type) {
	case *astpkg.NumberExpression:
		return e.Value, true
	case *astpkg.StringExpression:
		return e.Value, true
	case *astpkg.BoolExpression:
		return e.Value, true
	default:
		return nil, false
	}
}

func boolLiteral(expr astpkg.Expression) (bool, bool) {
	if b, ok := expr.(*astpkg.BoolExpression); ok {
		return b.Value, true
	}
	return false, false
}

func foldBinary(op token.TokenType, lexeme string, left, right any, span source.Span) astpkg.Expression {
	switch op {
	case token.PLUS:
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return &astpkg.NumberExpression{BaseNode: astpkg.BaseNode{NodeSpan: span}, Value: l + r}
			}
		}
		if _, ok := left.(string); ok {
			return &astpkg.StringExpression{BaseNode: astpkg.BaseNode{NodeSpan: span}, Value: fmt.Sprint(left) + fmt.Sprint(right)}
		}
		if _, ok := right.(string); ok {
			return &astpkg.StringExpression{BaseNode: astpkg.BaseNode{NodeSpan: span}, Value: fmt.Sprint(left) + fmt.Sprint(right)}
		}
	case token.MINUS:
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return &astpkg.NumberExpression{BaseNode: astpkg.BaseNode{NodeSpan: span}, Value: l - r}
			}
		}
	case token.STAR:
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return &astpkg.NumberExpression{BaseNode: astpkg.BaseNode{NodeSpan: span}, Value: l * r}
			}
		}
	case token.SLASH:
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok && r != 0 {
				return &astpkg.NumberExpression{BaseNode: astpkg.BaseNode{NodeSpan: span}, Value: l / r}
			}
		}
	case token.GREATER:
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return &astpkg.BoolExpression{BaseNode: astpkg.BaseNode{NodeSpan: span}, Value: l > r}
			}
		}
	case token.GREATER_EQUAL:
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return &astpkg.BoolExpression{BaseNode: astpkg.BaseNode{NodeSpan: span}, Value: l >= r}
			}
		}
	case token.LESS:
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return &astpkg.BoolExpression{BaseNode: astpkg.BaseNode{NodeSpan: span}, Value: l < r}
			}
		}
	case token.LESS_EQUAL:
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return &astpkg.BoolExpression{BaseNode: astpkg.BaseNode{NodeSpan: span}, Value: l <= r}
			}
		}
	case token.EQUAL_EQUAL:
		return &astpkg.BoolExpression{BaseNode: astpkg.BaseNode{NodeSpan: span}, Value: equalLiteral(left, right)}
	case token.BANG_EQUAL:
		return &astpkg.BoolExpression{BaseNode: astpkg.BaseNode{NodeSpan: span}, Value: !equalLiteral(left, right)}
	case token.AND_AND:
		if l, ok := left.(bool); ok {
			if r, ok := right.(bool); ok {
				return &astpkg.BoolExpression{BaseNode: astpkg.BaseNode{NodeSpan: span}, Value: l && r}
			}
		}
	case token.OR_OR:
		if l, ok := left.(bool); ok {
			if r, ok := right.(bool); ok {
				return &astpkg.BoolExpression{BaseNode: astpkg.BaseNode{NodeSpan: span}, Value: l || r}
			}
		}
	}
	_ = lexeme
	return nil
}

func equalLiteral(a, b any) bool {
	if af, ok := a.(float64); ok {
		if bf, ok := b.(float64); ok {
			return math.Abs(af-bf) < 1e-9
		}
	}
	return fmt.Sprintf("%#v", a) == fmt.Sprintf("%#v", b)
}

func isPure(expr astpkg.Expression) bool {
	switch e := expr.(type) {
	case *astpkg.NumberExpression, *astpkg.StringExpression, *astpkg.BoolExpression, *astpkg.VariableExpression:
		return true
	case *astpkg.GroupingExpression:
		return isPure(e.Expr)
	case *astpkg.UnaryExpression:
		return isPure(e.Right)
	case *astpkg.BinaryExpression:
		return isPure(e.Left) && isPure(e.Right)
	case *astpkg.ArrayExpression:
		for _, element := range e.Elements {
			if !isPure(element) {
				return false
			}
		}
		return true
	case *astpkg.IndexExpression:
		return isPure(e.Array) && isPure(e.Index)
	default:
		return false
	}
}
