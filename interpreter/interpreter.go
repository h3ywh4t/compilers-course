package interpreter

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	astpkg "compilerlabs/ast"
	rtenv "compilerlabs/runtime"
	"compilerlabs/source"
	"compilerlabs/token"
)

type Interpreter struct {
	environment *rtenv.Environment
	output      io.Writer
}

func New(output io.Writer) *Interpreter {
	if output == nil {
		output = io.Discard
	}
	return &Interpreter{environment: rtenv.NewEnvironment(nil), output: output}
}

func (i *Interpreter) Interpret(statements []astpkg.Statement) error {
	for _, stmt := range statements {
		if err := i.Execute(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) Execute(stmt astpkg.Statement) error {
	switch s := stmt.(type) {
	case *astpkg.PrintStatement:
		value, err := i.Eval(s.Expr)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(i.output, FormatValue(value))
		return err
	case *astpkg.VarStatement:
		var value any
		var err error
		if s.Initializer != nil {
			value, err = i.Eval(s.Initializer)
			if err != nil {
				return err
			}
		}
		i.environment.Define(s.Name, value)
		return nil
	case *astpkg.ExpressionStatement:
		_, err := i.Eval(s.Expr)
		return err
	case *astpkg.BlockStatement:
		return i.executeBlock(s.Statements, rtenv.NewEnvironment(i.environment))
	case *astpkg.IfStatement:
		cond, err := i.Eval(s.Condition)
		if err != nil {
			return err
		}
		if truthy(cond) {
			return i.Execute(s.ThenBranch)
		}
		if s.ElseBranch != nil {
			return i.Execute(s.ElseBranch)
		}
		return nil
	case *astpkg.WhileStatement:
		for {
			cond, err := i.Eval(s.Condition)
			if err != nil {
				return err
			}
			if !truthy(cond) {
				break
			}
			if err := i.Execute(s.Body); err != nil {
				return err
			}
		}
		return nil
	case *astpkg.FunctionStatement:
		i.environment.DefineFunction(s.Name, &rtenv.FunctionValue{Declaration: s, Closure: i.environment})
		return nil
	case *astpkg.ReturnStatement:
		var value any
		var err error
		if s.Value != nil {
			value, err = i.Eval(s.Value)
			if err != nil {
				return err
			}
		}
		return &returnSignal{value: value}
	default:
		return i.errorAt(stmt.Span(), fmt.Sprintf("неизвестная инструкция %T", stmt))
	}
}

func (i *Interpreter) executeBlock(statements []astpkg.Statement, env *rtenv.Environment) error {
	previous := i.environment
	i.environment = env
	defer func() { i.environment = previous }()
	for _, stmt := range statements {
		if err := i.Execute(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) Eval(expr astpkg.Expression) (any, error) {
	switch e := expr.(type) {
	case *astpkg.NumberExpression:
		return e.Value, nil
	case *astpkg.StringExpression:
		return e.Value, nil
	case *astpkg.BoolExpression:
		return e.Value, nil
	case *astpkg.VariableExpression:
		return i.environment.Get(e.Name)
	case *astpkg.GroupingExpression:
		return i.Eval(e.Expr)
	case *astpkg.ArrayExpression:
		values := make([]any, 0, len(e.Elements))
		for _, element := range e.Elements {
			v, err := i.Eval(element)
			if err != nil {
				return nil, err
			}
			values = append(values, v)
		}
		return values, nil
	case *astpkg.AssignExpression:
		value, err := i.Eval(e.Value)
		if err != nil {
			return nil, err
		}
		return value, i.environment.Assign(e.Name, value)
	case *astpkg.IndexExpression:
		return i.evalIndex(e.Array, e.Index, e.Span())
	case *astpkg.IndexAssignExpression:
		arr, idx, err := i.arrayAndIndex(e.Array, e.Index, e.Span())
		if err != nil {
			return nil, err
		}
		value, err := i.Eval(e.Value)
		if err != nil {
			return nil, err
		}
		arr[idx] = value
		return value, nil
	case *astpkg.UnaryExpression:
		right, err := i.Eval(e.Right)
		if err != nil {
			return nil, err
		}
		switch e.Operator.Type {
		case token.MINUS:
			return -right.(float64), nil
		case token.BANG:
			return !truthy(right), nil
		default:
			return nil, i.errorAt(e.Span(), "неизвестный унарный оператор")
		}
	case *astpkg.BinaryExpression:
		return i.evalBinary(e)
	case *astpkg.CallExpression:
		return i.evalCall(e)
	default:
		return nil, i.errorAt(expr.Span(), fmt.Sprintf("неизвестное выражение %T", expr))
	}
}

func (i *Interpreter) evalBinary(e *astpkg.BinaryExpression) (any, error) {
	if e.Operator.Type == token.OR_OR {
		left, err := i.Eval(e.Left)
		if err != nil {
			return nil, err
		}
		if truthy(left) {
			return true, nil
		}
		right, err := i.Eval(e.Right)
		if err != nil {
			return nil, err
		}
		return truthy(right), nil
	}
	if e.Operator.Type == token.AND_AND {
		left, err := i.Eval(e.Left)
		if err != nil {
			return nil, err
		}
		if !truthy(left) {
			return false, nil
		}
		right, err := i.Eval(e.Right)
		if err != nil {
			return nil, err
		}
		return truthy(right), nil
	}
	left, err := i.Eval(e.Left)
	if err != nil {
		return nil, err
	}
	right, err := i.Eval(e.Right)
	if err != nil {
		return nil, err
	}
	switch e.Operator.Type {
	case token.PLUS:
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				return l + r, nil
			}
		}
		if _, ok := left.(string); ok {
			return FormatValue(left) + FormatValue(right), nil
		}
		if _, ok := right.(string); ok {
			return FormatValue(left) + FormatValue(right), nil
		}
	case token.MINUS:
		return left.(float64) - right.(float64), nil
	case token.STAR:
		return left.(float64) * right.(float64), nil
	case token.SLASH:
		if right.(float64) == 0 {
			return nil, i.errorAt(e.Span(), "деление на ноль")
		}
		return left.(float64) / right.(float64), nil
	case token.GREATER:
		return left.(float64) > right.(float64), nil
	case token.GREATER_EQUAL:
		return left.(float64) >= right.(float64), nil
	case token.LESS:
		return left.(float64) < right.(float64), nil
	case token.LESS_EQUAL:
		return left.(float64) <= right.(float64), nil
	case token.EQUAL_EQUAL:
		return equal(left, right), nil
	case token.BANG_EQUAL:
		return !equal(left, right), nil
	}
	return nil, i.errorAt(e.Span(), "неизвестный бинарный оператор")
}

func (i *Interpreter) evalCall(e *astpkg.CallExpression) (any, error) {
	fn, err := i.environment.GetFunction(e.CalleeName)
	if err != nil {
		return nil, err
	}
	args := make([]any, 0, len(e.Arguments))
	for _, arg := range e.Arguments {
		value, err := i.Eval(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, value)
	}
	callEnv := rtenv.NewEnvironment(fn.Closure)
	for idx, param := range fn.Declaration.Parameters {
		var arg any
		if idx < len(args) {
			arg = args[idx]
		}
		callEnv.Define(param, arg)
	}
	previous := i.environment
	i.environment = callEnv
	defer func() { i.environment = previous }()
	for _, stmt := range fn.Declaration.Body.Statements {
		if err := i.Execute(stmt); err != nil {
			if ret, ok := err.(*returnSignal); ok {
				return ret.value, nil
			}
			return nil, err
		}
	}
	return nil, nil
}

func (i *Interpreter) evalIndex(arrayExpr, indexExpr astpkg.Expression, span source.Span) (any, error) {
	arr, idx, err := i.arrayAndIndex(arrayExpr, indexExpr, span)
	if err != nil {
		return nil, err
	}
	return arr[idx], nil
}

func (i *Interpreter) arrayAndIndex(arrayExpr, indexExpr astpkg.Expression, span source.Span) ([]any, int, error) {
	arrayValue, err := i.Eval(arrayExpr)
	if err != nil {
		return nil, 0, err
	}
	arr, ok := arrayValue.([]any)
	if !ok {
		return nil, 0, i.errorAt(span, "значение не является массивом")
	}
	indexValue, err := i.Eval(indexExpr)
	if err != nil {
		return nil, 0, err
	}
	num, ok := indexValue.(float64)
	if !ok {
		return nil, 0, i.errorAt(span, "индекс массива должен быть числом")
	}
	if math.Trunc(num) != num {
		return nil, 0, i.errorAt(span, "индекс массива должен быть целым числом")
	}
	idx := int(num)
	if idx < 0 || idx >= len(arr) {
		return nil, 0, i.errorAt(span, fmt.Sprintf("индекс %d вне границ массива длины %d", idx, len(arr)))
	}
	return arr, idx, nil
}

func (i *Interpreter) errorAt(span source.Span, message string) error {
	return &RuntimeError{Message: message, Span: span}
}

func truthy(v any) bool   { b, ok := v.(bool); return ok && b }
func equal(a, b any) bool { return fmt.Sprintf("%#v", a) == fmt.Sprintf("%#v", b) }

func FormatValue(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case float64:
		if math.Trunc(x) == x {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case []any:
		parts := make([]string, len(x))
		for idx, item := range x {
			parts[idx] = FormatValue(item)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return fmt.Sprint(v)
	}
}
