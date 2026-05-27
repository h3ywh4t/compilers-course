package interpreter

import (
	"bytes"
	"compilerlabs/lexer"
	"compilerlabs/parser"
	"compilerlabs/semantic"
	"testing"
)

func runSource(t *testing.T, src string) (string, error) {
	t.Helper()
	tokens, err := lexer.New(src).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	program, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if errs := semantic.New().Analyze(program); len(errs) != 0 {
		t.Fatalf("semantic errors: %v", errs)
	}
	var out bytes.Buffer
	err = New(&out).Interpret(program)
	return out.String(), err
}

func TestInterpreterFunctionArray(t *testing.T) {
	out, err := runSource(t, `fun f(x){ return x * 2; } var a=[1,2]; a[1]=f(10); print a[1];`)
	if err != nil {
		t.Fatalf("runtime error: %v", err)
	}
	if out != "20\n" {
		t.Fatalf("got output %q", out)
	}
}
