package lexer

import (
	"compilerlabs/token"
	"testing"
)

func TestScanTokensExtended(t *testing.T) {
	src := `fun f(x) { var a = [1, 2.5]; print "ok" + true; }`
	l := New(src)
	tokens, err := l.ScanTokens()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	want := []token.TokenType{token.FUN, token.IDENTIFIER, token.LEFT_PAREN, token.IDENTIFIER, token.RIGHT_PAREN, token.LEFT_BRACE, token.VAR, token.IDENTIFIER, token.EQUAL, token.LEFT_BRACKET, token.NUMBER, token.COMMA, token.NUMBER, token.RIGHT_BRACKET, token.SEMICOLON, token.PRINT, token.STRING, token.PLUS, token.TRUE, token.SEMICOLON, token.RIGHT_BRACE, token.EOF}
	if len(tokens) != len(want) {
		t.Fatalf("token count got %d want %d", len(tokens), len(want))
	}
	for i := range want {
		if tokens[i].Type != want[i] {
			t.Fatalf("token[%d] got %s want %s", i, tokens[i].Type, want[i])
		}
	}
}

func TestScanTokensInvalidCharacter(t *testing.T) {
	_, err := New("var x = @;").ScanTokens()
	if err == nil {
		t.Fatal("expected lexer error")
	}
}
