package token

import (
	"fmt"

	"compilerlabs/source"
)

type TokenType string

const (
	LEFT_PAREN    TokenType = "LEFT_PAREN"
	RIGHT_PAREN   TokenType = "RIGHT_PAREN"
	LEFT_BRACE    TokenType = "LEFT_BRACE"
	RIGHT_BRACE   TokenType = "RIGHT_BRACE"
	LEFT_BRACKET  TokenType = "LEFT_BRACKET"
	RIGHT_BRACKET TokenType = "RIGHT_BRACKET"
	COMMA         TokenType = "COMMA"
	COLON         TokenType = "COLON"
	SEMICOLON     TokenType = "SEMICOLON"
	PLUS          TokenType = "PLUS"
	MINUS         TokenType = "MINUS"
	STAR          TokenType = "STAR"
	SLASH         TokenType = "SLASH"
	BANG          TokenType = "BANG"
	EQUAL         TokenType = "EQUAL"
	LESS          TokenType = "LESS"
	GREATER       TokenType = "GREATER"

	BANG_EQUAL    TokenType = "BANG_EQUAL"
	EQUAL_EQUAL   TokenType = "EQUAL_EQUAL"
	LESS_EQUAL    TokenType = "LESS_EQUAL"
	GREATER_EQUAL TokenType = "GREATER_EQUAL"
	AND_AND       TokenType = "AND_AND"
	OR_OR         TokenType = "OR_OR"

	IDENTIFIER TokenType = "IDENTIFIER"
	NUMBER     TokenType = "NUMBER"
	STRING     TokenType = "STRING"

	VAR    TokenType = "VAR"
	PRINT  TokenType = "PRINT"
	IF     TokenType = "IF"
	ELSE   TokenType = "ELSE"
	WHILE  TokenType = "WHILE"
	FUN    TokenType = "FUN"
	RETURN TokenType = "RETURN"
	TRUE   TokenType = "TRUE"
	FALSE  TokenType = "FALSE"

	EOF TokenType = "EOF"
)

type Token struct {
	Type    TokenType
	Lexeme  string
	Literal any
	Span    source.Span
}

func (t Token) String() string {
	if t.Literal != nil {
		return fmt.Sprintf("%-15s lexeme=%q literal=%v [%s]", t.Type, t.Lexeme, t.Literal, t.Span)
	}
	return fmt.Sprintf("%-15s lexeme=%q [%s]", t.Type, t.Lexeme, t.Span)
}

var keywords = map[string]TokenType{
	"var":    VAR,
	"print":  PRINT,
	"if":     IF,
	"else":   ELSE,
	"while":  WHILE,
	"fun":    FUN,
	"return": RETURN,
	"true":   TRUE,
	"false":  FALSE,
}

func LookupIdentifier(lexeme string) TokenType {
	if tt, ok := keywords[lexeme]; ok {
		return tt
	}
	return IDENTIFIER
}
