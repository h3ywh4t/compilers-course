package parser

import "compilerlabs/token"

func (p *Parser) synchronize() {
	p.Advance()
	for !p.isAtEnd() {
		if p.Previous().Type == token.SEMICOLON {
			return
		}
		switch p.Peek().Type {
		case token.VAR, token.FUN, token.IF, token.WHILE, token.PRINT, token.RETURN:
			return
		}
		p.Advance()
	}
}
