package lexer

import (
	"fmt"
	"strconv"
	"strings"

	"compilerlabs/source"
	"compilerlabs/token"
)

type Lexer struct {
	source []rune
	tokens []token.Token

	start   int
	current int

	line   int
	column int

	startPos source.Position
}

func New(input string) *Lexer {
	return &Lexer{source: []rune(input), line: 1, column: 1}
}

func (l *Lexer) ScanTokens() ([]token.Token, error) {
	for !l.isAtEnd() {
		l.start = l.current
		l.startPos = l.currentPos()
		if err := l.scanToken(); err != nil {
			return nil, err
		}
	}
	eofPos := l.currentPos()
	l.tokens = append(l.tokens, token.Token{Type: token.EOF, Span: source.Span{Start: eofPos, End: eofPos}})
	return l.tokens, nil
}

func (l *Lexer) scanToken() error {
	c := l.advance()
	switch c {
	case ' ', '\r', '\t', '\n':
		return nil
	case '(':
		l.addToken(token.LEFT_PAREN, nil)
	case ')':
		l.addToken(token.RIGHT_PAREN, nil)
	case '{':
		l.addToken(token.LEFT_BRACE, nil)
	case '}':
		l.addToken(token.RIGHT_BRACE, nil)
	case '[':
		l.addToken(token.LEFT_BRACKET, nil)
	case ']':
		l.addToken(token.RIGHT_BRACKET, nil)
	case ',':
		l.addToken(token.COMMA, nil)
	case ':':
		l.addToken(token.COLON, nil)
	case ';':
		l.addToken(token.SEMICOLON, nil)
	case '+':
		l.addToken(token.PLUS, nil)
	case '-':
		l.addToken(token.MINUS, nil)
	case '*':
		l.addToken(token.STAR, nil)
	case '!':
		if l.match('=') {
			l.addToken(token.BANG_EQUAL, nil)
		} else {
			l.addToken(token.BANG, nil)
		}
	case '=':
		if l.match('=') {
			l.addToken(token.EQUAL_EQUAL, nil)
		} else {
			l.addToken(token.EQUAL, nil)
		}
	case '<':
		if l.match('=') {
			l.addToken(token.LESS_EQUAL, nil)
		} else {
			l.addToken(token.LESS, nil)
		}
	case '>':
		if l.match('=') {
			l.addToken(token.GREATER_EQUAL, nil)
		} else {
			l.addToken(token.GREATER, nil)
		}
	case '&':
		if l.match('&') {
			l.addToken(token.AND_AND, nil)
		} else {
			return l.errorf("ожидался второй символ '&' для оператора &&")
		}
	case '|':
		if l.match('|') {
			l.addToken(token.OR_OR, nil)
		} else {
			return l.errorf("ожидался второй символ '|' для оператора ||")
		}
	case '/':
		if l.match('/') {
			for l.peek() != '\n' && !l.isAtEnd() {
				l.advance()
			}
		} else if l.match('*') {
			return l.blockComment()
		} else {
			l.addToken(token.SLASH, nil)
		}
	case '"':
		return l.string()
	default:
		if isDigit(c) {
			return l.number()
		}
		if isAlpha(c) {
			return l.identifier()
		}
		return l.errorf(fmt.Sprintf("неожиданный символ %q", c))
	}
	return nil
}

func (l *Lexer) blockComment() error {
	for !l.isAtEnd() {
		if l.peek() == '*' && l.peekNext() == '/' {
			l.advance()
			l.advance()
			return nil
		}
		l.advance()
	}
	return l.errorf("незакрытый многострочный комментарий")
}

func (l *Lexer) string() error {
	var b strings.Builder
	for !l.isAtEnd() && l.peek() != '"' {
		ch := l.advance()
		if ch == '\\' && !l.isAtEnd() {
			next := l.advance()
			switch next {
			case 'n':
				b.WriteRune('\n')
			case 't':
				b.WriteRune('\t')
			case '"':
				b.WriteRune('"')
			case '\\':
				b.WriteRune('\\')
			default:
				b.WriteRune(next)
			}
			continue
		}
		b.WriteRune(ch)
	}
	if l.isAtEnd() {
		return l.errorf("незакрытая строка")
	}
	l.advance() // closing quote
	l.tokens = append(l.tokens, token.Token{Type: token.STRING, Lexeme: string(l.source[l.start:l.current]), Literal: b.String(), Span: l.currentSpan()})
	return nil
}

func (l *Lexer) number() error {
	for isDigit(l.peek()) {
		l.advance()
	}
	if l.peek() == '.' && isDigit(l.peekNext()) {
		l.advance()
		for isDigit(l.peek()) {
			l.advance()
		}
	}
	lexeme := string(l.source[l.start:l.current])
	value, err := strconv.ParseFloat(lexeme, 64)
	if err != nil {
		return l.errorf("не удалось распознать число")
	}
	l.tokens = append(l.tokens, token.Token{Type: token.NUMBER, Lexeme: lexeme, Literal: value, Span: l.currentSpan()})
	return nil
}

func (l *Lexer) identifier() error {
	for isAlphaNumeric(l.peek()) {
		l.advance()
	}
	lexeme := string(l.source[l.start:l.current])
	l.tokens = append(l.tokens, token.Token{Type: token.LookupIdentifier(lexeme), Lexeme: lexeme, Span: l.currentSpan()})
	return nil
}

func (l *Lexer) addToken(tt token.TokenType, literal any) {
	l.tokens = append(l.tokens, token.Token{Type: tt, Lexeme: string(l.source[l.start:l.current]), Literal: literal, Span: l.currentSpan()})
}

func (l *Lexer) isAtEnd() bool { return l.current >= len(l.source) }

func (l *Lexer) advance() rune {
	ch := l.source[l.current]
	l.current++
	if ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return ch
}

func (l *Lexer) match(expected rune) bool {
	if l.isAtEnd() || l.source[l.current] != expected {
		return false
	}
	l.advance()
	return true
}
func (l *Lexer) peek() rune {
	if l.isAtEnd() {
		return '\000'
	}
	return l.source[l.current]
}
func (l *Lexer) peekNext() rune {
	if l.current+1 >= len(l.source) {
		return '\000'
	}
	return l.source[l.current+1]
}
func (l *Lexer) currentPos() source.Position { return source.Position{Line: l.line, Column: l.column} }
func (l *Lexer) currentSpan() source.Span    { return source.Span{Start: l.startPos, End: l.currentPos()} }
func (l *Lexer) errorf(message string) error {
	return fmt.Errorf("lexer error at %s: %s", l.startPos, message)
}
func isDigit(r rune) bool        { return r >= '0' && r <= '9' }
func isAlpha(r rune) bool        { return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' }
func isAlphaNumeric(r rune) bool { return isAlpha(r) || isDigit(r) }
