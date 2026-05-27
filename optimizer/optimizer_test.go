package optimizer

import (
	astpkg "compilerlabs/ast"
	"compilerlabs/lexer"
	"compilerlabs/parser"
	"testing"
)

func parseOpt(t *testing.T, src string) []astpkg.Statement {
	t.Helper()
	tokens, err := lexer.New(src).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	program, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	return New().Optimize(program)
}

func TestConstantFoldingAndDCE(t *testing.T) {
	program := parseOpt(t, `if (false) { print 1; } else { print 2 + 3; }`)
	if len(program) != 1 {
		t.Fatalf("expected one statement, got %d", len(program))
	}
	ps, ok := program[0].(*astpkg.BlockStatement)
	if !ok {
		t.Fatalf("expected block from else branch, got %T", program[0])
	}
	printStmt := ps.Statements[0].(*astpkg.PrintStatement)
	if n, ok := printStmt.Expr.(*astpkg.NumberExpression); !ok || n.Value != 5 {
		t.Fatalf("constant folding failed: %#v", printStmt.Expr)
	}
}
