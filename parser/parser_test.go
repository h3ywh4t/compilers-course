package parser

import (
	astpkg "compilerlabs/ast"
	"compilerlabs/lexer"
	"testing"
)

func parseSource(t *testing.T, src string) ([]astpkg.Statement, error) {
	t.Helper()
	tokens, err := lexer.New(src).ScanTokens()
	if err != nil {
		t.Fatalf("lexer: %v", err)
	}
	return New(tokens).Parse()
}

func TestParseFunctionArrayAndIndexAssign(t *testing.T) {
	program, err := parseSource(t, `fun f(x) { return x * 2; } var a = [1,2,3]; a[0] = f(10); print a[0];`)
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}
	if len(program) != 4 {
		t.Fatalf("expected 4 statements, got %d", len(program))
	}
	if _, ok := program[0].(*astpkg.FunctionStatement); !ok {
		t.Fatalf("expected function statement")
	}
}

func TestParseMissingSemicolon(t *testing.T) {
	_, err := parseSource(t, `var x = 10`)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestParseExplicitTypeAnnotations(t *testing.T) {
	program, err := parseSource(t, `
var x : number = 10;
var names : string[] = ["a", "b"];
fun add(a : number, b : number) : number {
	return a + b;
}
`)
	if err != nil {
		t.Fatalf("unexpected parser error: %v", err)
	}
	if len(program) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(program))
	}
}
