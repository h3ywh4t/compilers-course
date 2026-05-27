package llmr

import (
	"compilerlabs/lexer"
	"compilerlabs/parser"
	"strings"
	"testing"
)

func TestTranslate(t *testing.T) {
	tokens, err := lexer.New(`var x = 1 + 2; print x;`).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	program, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	ir := NewTranslator().Translate(program).String()
	if !strings.Contains(ir, "add") || !strings.Contains(ir, "print") {
		t.Fatalf("unexpected ir: %s", ir)
	}
}
