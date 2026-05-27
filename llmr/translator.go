package llmr

import (
	"fmt"
	"strconv"

	astpkg "compilerlabs/ast"
	"compilerlabs/token"
)

type Translator struct {
	tempID  int
	labelID int
	program Program
}

func NewTranslator() *Translator { return &Translator{} }
func (t *Translator) Translate(statements []astpkg.Statement) Program {
	for _, stmt := range statements {
		t.statement(stmt)
	}
	return t.program
}
func (t *Translator) emit(op, result string, args ...string) {
	t.program.Instructions = append(t.program.Instructions, Instruction{Op: op, Result: result, Args: args})
}
func (t *Translator) temp() string { t.tempID++; return fmt.Sprintf("%%t%d", t.tempID) }
func (t *Translator) label(prefix string) string {
	t.labelID++
	return fmt.Sprintf("%s_%d", prefix, t.labelID)
}

func (t *Translator) statement(stmt astpkg.Statement) {
	switch s := stmt.(type) {
	case *astpkg.VarStatement:
		if s.Initializer != nil {
			v := t.expression(s.Initializer)
			t.emit("store", "", s.Name, v)
		} else {
			t.emit("store", "", s.Name, "nil")
		}
	case *astpkg.PrintStatement:
		v := t.expression(s.Expr)
		t.emit("print", "", v)
	case *astpkg.ExpressionStatement:
		t.expression(s.Expr)
	case *astpkg.BlockStatement:
		for _, inner := range s.Statements {
			t.statement(inner)
		}
	case *astpkg.IfStatement:
		elseLabel := t.label("else")
		endLabel := t.label("endif")
		cond := t.expression(s.Condition)
		t.emit("jump_if_false", "", cond, elseLabel)
		t.statement(s.ThenBranch)
		t.emit("jump", "", endLabel)
		t.emit("label", "", elseLabel)
		if s.ElseBranch != nil {
			t.statement(s.ElseBranch)
		}
		t.emit("label", "", endLabel)
	case *astpkg.WhileStatement:
		start := t.label("while_start")
		end := t.label("while_end")
		t.emit("label", "", start)
		cond := t.expression(s.Condition)
		t.emit("jump_if_false", "", cond, end)
		t.statement(s.Body)
		t.emit("jump", "", start)
		t.emit("label", "", end)
	case *astpkg.FunctionStatement:
		args := append([]string{s.Name}, s.Parameters...)
		t.emit("func", "", args...)
		for _, inner := range s.Body.Statements {
			t.statement(inner)
		}
		t.emit("end_func", "", s.Name)
	case *astpkg.ReturnStatement:
		if s.Value != nil {
			v := t.expression(s.Value)
			t.emit("return", "", v)
		} else {
			t.emit("return", "", "nil")
		}
	}
}

func (t *Translator) expression(expr astpkg.Expression) string {
	switch e := expr.(type) {
	case *astpkg.NumberExpression:
		out := t.temp()
		t.emit("const", out, strconv.FormatFloat(e.Value, 'f', -1, 64))
		return out
	case *astpkg.StringExpression:
		out := t.temp()
		t.emit("const", out, strconv.Quote(e.Value))
		return out
	case *astpkg.BoolExpression:
		out := t.temp()
		t.emit("const", out, strconv.FormatBool(e.Value))
		return out
	case *astpkg.VariableExpression:
		out := t.temp()
		t.emit("load", out, e.Name)
		return out
	case *astpkg.GroupingExpression:
		return t.expression(e.Expr)
	case *astpkg.ArrayExpression:
		args := make([]string, 0, len(e.Elements))
		for _, el := range e.Elements {
			args = append(args, t.expression(el))
		}
		out := t.temp()
		t.emit("array", out, args...)
		return out
	case *astpkg.IndexExpression:
		arr := t.expression(e.Array)
		idx := t.expression(e.Index)
		out := t.temp()
		t.emit("index_load", out, arr, idx)
		return out
	case *astpkg.AssignExpression:
		v := t.expression(e.Value)
		t.emit("store", "", e.Name, v)
		return v
	case *astpkg.IndexAssignExpression:
		arr := t.expression(e.Array)
		idx := t.expression(e.Index)
		v := t.expression(e.Value)
		t.emit("index_store", "", arr, idx, v)
		return v
	case *astpkg.UnaryExpression:
		r := t.expression(e.Right)
		out := t.temp()
		t.emit(unaryOp(e.Operator.Type), out, r)
		return out
	case *astpkg.BinaryExpression:
		l := t.expression(e.Left)
		r := t.expression(e.Right)
		out := t.temp()
		t.emit(binaryOp(e.Operator.Type), out, l, r)
		return out
	case *astpkg.CallExpression:
		args := []string{e.CalleeName}
		for _, arg := range e.Arguments {
			args = append(args, t.expression(arg))
		}
		out := t.temp()
		t.emit("call", out, args...)
		return out
	default:
		out := t.temp()
		t.emit("unknown", out)
		return out
	}
}

func unaryOp(tt token.TokenType) string {
	if tt == token.MINUS {
		return "neg"
	}
	if tt == token.BANG {
		return "not"
	}
	return "unary"
}
func binaryOp(tt token.TokenType) string {
	switch tt {
	case token.PLUS:
		return "add"
	case token.MINUS:
		return "sub"
	case token.STAR:
		return "mul"
	case token.SLASH:
		return "div"
	case token.LESS:
		return "lt"
	case token.LESS_EQUAL:
		return "le"
	case token.GREATER:
		return "gt"
	case token.GREATER_EQUAL:
		return "ge"
	case token.EQUAL_EQUAL:
		return "eq"
	case token.BANG_EQUAL:
		return "ne"
	case token.AND_AND:
		return "and"
	case token.OR_OR:
		return "or"
	default:
		return "bin"
	}
}
