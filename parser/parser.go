package parser

import (
	"fmt"

	astpkg "compilerlabs/ast"
	"compilerlabs/source"
	"compilerlabs/token"
	"compilerlabs/types"
)

type Parser struct {
	tokens  []token.Token
	current int
}

func New(tokens []token.Token) *Parser { return &Parser{tokens: tokens} }

func (p *Parser) Parse() ([]astpkg.Statement, error) {
	statements := make([]astpkg.Statement, 0)
	for !p.isAtEnd() {
		stmt, err := p.parseDeclaration()
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}
	return statements, nil
}

func (p *Parser) parseDeclaration() (astpkg.Statement, error) {
	if p.Match(token.FUN) {
		return p.parseFunctionDeclaration(p.Previous())
	}
	if p.Match(token.VAR) {
		return p.parseVarDeclaration(p.Previous())
	}
	return p.parseStatement()
}

func (p *Parser) parseFunctionDeclaration(funTok token.Token) (astpkg.Statement, error) {
	name, err := p.Consume(token.IDENTIFIER, "ожидается имя функции")
	if err != nil {
		return nil, err
	}
	if _, err = p.Consume(token.LEFT_PAREN, "ожидается '(' после имени функции"); err != nil {
		return nil, err
	}
	params := make([]string, 0)
	paramTypes := make([]*types.Type, 0)
	if !p.Check(token.RIGHT_PAREN) {
		for {
			param, err := p.Consume(token.IDENTIFIER, "ожидается имя параметра")
			if err != nil {
				return nil, err
			}
			paramType, err := p.parseOptionalTypeAnnotation()
			if err != nil {
				return nil, err
			}
			params = append(params, param.Lexeme)
			paramTypes = append(paramTypes, paramType)
			if !p.Match(token.COMMA) {
				break
			}
		}
	}
	if _, err = p.Consume(token.RIGHT_PAREN, "ожидается ')' после параметров функции"); err != nil {
		return nil, err
	}
	returnType, err := p.parseOptionalTypeAnnotation()
	if err != nil {
		return nil, err
	}
	open, err := p.Consume(token.LEFT_BRACE, "ожидается '{' перед телом функции")
	if err != nil {
		return nil, err
	}
	bodyStmt, err := p.parseBlockStatement(open)
	if err != nil {
		return nil, err
	}
	body := bodyStmt.(*astpkg.BlockStatement)
	return &astpkg.FunctionStatement{
		BaseNode:       astpkg.BaseNode{NodeSpan: source.Merge(funTok.Span, body.Span())},
		Name:           name.Lexeme,
		Parameters:     params,
		ParameterTypes: paramTypes,
		ReturnType:     returnType,
		Body:           body,
	}, nil
}

func (p *Parser) parseVarDeclaration(varTok token.Token) (astpkg.Statement, error) {
	nameTok, err := p.Consume(token.IDENTIFIER, "ожидается имя переменной")
	if err != nil {
		return nil, err
	}
	typeAnnotation, err := p.parseOptionalTypeAnnotation()
	if err != nil {
		return nil, err
	}
	var initializer astpkg.Expression
	if p.Match(token.EQUAL) {
		initializer, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}
	semi, err := p.Consume(token.SEMICOLON, "ожидается ';' после объявления переменной")
	if err != nil {
		return nil, err
	}
	span := source.Merge(varTok.Span, semi.Span)
	return &astpkg.VarStatement{BaseNode: astpkg.BaseNode{NodeSpan: span}, Name: nameTok.Lexeme, TypeAnnotation: typeAnnotation, Initializer: initializer}, nil
}

func (p *Parser) parseOptionalTypeAnnotation() (*types.Type, error) {
	if !p.Match(token.COLON) {
		return nil, nil
	}
	t, err := p.parseType()
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (p *Parser) parseType() (types.Type, error) {
	name, err := p.Consume(token.IDENTIFIER, "ожидается тип: number, string или bool")
	if err != nil {
		return types.Unknown, err
	}
	var out types.Type
	switch name.Lexeme {
	case "number":
		out = types.Number
	case "string":
		out = types.String
	case "bool":
		out = types.Bool
	default:
		return types.Unknown, p.errorAt(name, fmt.Sprintf("неизвестный тип '%s'", name.Lexeme))
	}
	for p.Match(token.LEFT_BRACKET) {
		if _, err := p.Consume(token.RIGHT_BRACKET, "ожидается ']' после [] в типе массива"); err != nil {
			return types.Unknown, err
		}
		out = types.ArrayOf(out)
	}
	return out, nil
}

func (p *Parser) parseStatement() (astpkg.Statement, error) {
	if p.Match(token.PRINT) {
		return p.parsePrintStatement(p.Previous())
	}
	if p.Match(token.IF) {
		return p.parseIfStatement(p.Previous())
	}
	if p.Match(token.WHILE) {
		return p.parseWhileStatement(p.Previous())
	}
	if p.Match(token.RETURN) {
		return p.parseReturnStatement(p.Previous())
	}
	if p.Match(token.LEFT_BRACE) {
		return p.parseBlockStatement(p.Previous())
	}
	return p.parseExpressionStatement()
}

func (p *Parser) parsePrintStatement(printTok token.Token) (astpkg.Statement, error) {
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	semi, err := p.Consume(token.SEMICOLON, "ожидается ';' после print")
	if err != nil {
		return nil, err
	}
	return &astpkg.PrintStatement{BaseNode: astpkg.BaseNode{NodeSpan: source.Merge(printTok.Span, semi.Span)}, Expr: expr}, nil
}

func (p *Parser) parseReturnStatement(returnTok token.Token) (astpkg.Statement, error) {
	var value astpkg.Expression
	var err error
	if !p.Check(token.SEMICOLON) {
		value, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}
	semi, err := p.Consume(token.SEMICOLON, "ожидается ';' после return")
	if err != nil {
		return nil, err
	}
	return &astpkg.ReturnStatement{BaseNode: astpkg.BaseNode{NodeSpan: source.Merge(returnTok.Span, semi.Span)}, Value: value}, nil
}

func (p *Parser) parseIfStatement(ifTok token.Token) (astpkg.Statement, error) {
	if _, err := p.Consume(token.LEFT_PAREN, "ожидается '(' после if"); err != nil {
		return nil, err
	}
	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if _, err = p.Consume(token.RIGHT_PAREN, "ожидается ')' после условия if"); err != nil {
		return nil, err
	}
	thenBranch, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	endSpan := thenBranch.Span()
	var elseBranch astpkg.Statement
	if p.Match(token.ELSE) {
		elseBranch, err = p.parseStatement()
		if err != nil {
			return nil, err
		}
		endSpan = elseBranch.Span()
	}
	return &astpkg.IfStatement{BaseNode: astpkg.BaseNode{NodeSpan: source.Merge(ifTok.Span, endSpan)}, Condition: condition, ThenBranch: thenBranch, ElseBranch: elseBranch}, nil
}

func (p *Parser) parseWhileStatement(whileTok token.Token) (astpkg.Statement, error) {
	if _, err := p.Consume(token.LEFT_PAREN, "ожидается '(' после while"); err != nil {
		return nil, err
	}
	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if _, err = p.Consume(token.RIGHT_PAREN, "ожидается ')' после условия while"); err != nil {
		return nil, err
	}
	body, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	return &astpkg.WhileStatement{BaseNode: astpkg.BaseNode{NodeSpan: source.Merge(whileTok.Span, body.Span())}, Condition: condition, Body: body}, nil
}

func (p *Parser) parseBlockStatement(open token.Token) (astpkg.Statement, error) {
	stmts, closeSpan, err := p.parseBlockContents()
	if err != nil {
		return nil, err
	}
	return &astpkg.BlockStatement{BaseNode: astpkg.BaseNode{NodeSpan: source.Merge(open.Span, closeSpan)}, Statements: stmts}, nil
}

func (p *Parser) parseBlockContents() ([]astpkg.Statement, source.Span, error) {
	stmts := make([]astpkg.Statement, 0)
	for !p.Check(token.RIGHT_BRACE) && !p.isAtEnd() {
		stmt, err := p.parseDeclaration()
		if err != nil {
			return nil, source.Span{}, err
		}
		stmts = append(stmts, stmt)
	}
	close, err := p.Consume(token.RIGHT_BRACE, "ожидается '}' после блока")
	if err != nil {
		return nil, source.Span{}, err
	}
	return stmts, close.Span, nil
}

func (p *Parser) parseExpressionStatement() (astpkg.Statement, error) {
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	semi, err := p.Consume(token.SEMICOLON, "ожидается ';' после выражения")
	if err != nil {
		return nil, err
	}
	return &astpkg.ExpressionStatement{BaseNode: astpkg.BaseNode{NodeSpan: source.Merge(expr.Span(), semi.Span)}, Expr: expr}, nil
}

func (p *Parser) parseExpression() (astpkg.Expression, error) { return p.parseAssignment() }

func (p *Parser) parseAssignment() (astpkg.Expression, error) {
	expr, err := p.parseLogicalOr()
	if err != nil {
		return nil, err
	}
	if p.Match(token.EQUAL) {
		eq := p.Previous()
		value, err := p.parseAssignment()
		if err != nil {
			return nil, err
		}
		switch target := expr.(type) {
		case *astpkg.VariableExpression:
			return &astpkg.AssignExpression{BaseNode: astpkg.BaseNode{NodeSpan: source.Merge(expr.Span(), value.Span())}, Name: target.Name, Value: value}, nil
		case *astpkg.IndexExpression:
			return &astpkg.IndexAssignExpression{BaseNode: astpkg.BaseNode{NodeSpan: source.Merge(expr.Span(), value.Span())}, Array: target.Array, Index: target.Index, Value: value}, nil
		default:
			return nil, p.errorAt(eq, "некорректная цель присваивания")
		}
	}
	return expr, nil
}

func (p *Parser) parseLogicalOr() (astpkg.Expression, error) {
	return p.parseBinary(p.parseLogicalAnd, token.OR_OR)
}
func (p *Parser) parseLogicalAnd() (astpkg.Expression, error) {
	return p.parseBinary(p.parseEquality, token.AND_AND)
}
func (p *Parser) parseEquality() (astpkg.Expression, error) {
	return p.parseBinary(p.parseComparison, token.EQUAL_EQUAL, token.BANG_EQUAL)
}
func (p *Parser) parseComparison() (astpkg.Expression, error) {
	return p.parseBinary(p.parseTerm, token.LESS, token.LESS_EQUAL, token.GREATER, token.GREATER_EQUAL)
}
func (p *Parser) parseTerm() (astpkg.Expression, error) {
	return p.parseBinary(p.parseFactor, token.PLUS, token.MINUS)
}
func (p *Parser) parseFactor() (astpkg.Expression, error) {
	return p.parseBinary(p.parseUnary, token.STAR, token.SLASH)
}

func (p *Parser) parseBinary(next func() (astpkg.Expression, error), types ...token.TokenType) (astpkg.Expression, error) {
	expr, err := next()
	if err != nil {
		return nil, err
	}
	for p.Match(types...) {
		op := p.Previous()
		right, err := next()
		if err != nil {
			return nil, err
		}
		expr = &astpkg.BinaryExpression{BaseNode: astpkg.BaseNode{NodeSpan: source.Merge(expr.Span(), right.Span())}, Left: expr, Operator: op, Right: right}
	}
	return expr, nil
}

func (p *Parser) parseUnary() (astpkg.Expression, error) {
	if p.Match(token.BANG, token.MINUS) {
		op := p.Previous()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &astpkg.UnaryExpression{BaseNode: astpkg.BaseNode{NodeSpan: source.Merge(op.Span, right.Span())}, Operator: op, Right: right}, nil
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() (astpkg.Expression, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for {
		if p.Match(token.LEFT_PAREN) {
			open := p.Previous()
			args, close, err := p.finishArguments(token.RIGHT_PAREN, "ожидается ')' после аргументов функции")
			if err != nil {
				return nil, err
			}
			v, ok := expr.(*astpkg.VariableExpression)
			if !ok {
				return nil, p.errorAt(open, "вызов функции возможен только по имени функции")
			}
			expr = &astpkg.CallExpression{BaseNode: astpkg.BaseNode{NodeSpan: source.Merge(expr.Span(), close.Span)}, CalleeName: v.Name, Arguments: args}
			continue
		}
		if p.Match(token.LEFT_BRACKET) {
			index, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			close, err := p.Consume(token.RIGHT_BRACKET, "ожидается ']' после индекса")
			if err != nil {
				return nil, err
			}
			expr = &astpkg.IndexExpression{BaseNode: astpkg.BaseNode{NodeSpan: source.Merge(expr.Span(), close.Span)}, Array: expr, Index: index}
			continue
		}
		break
	}
	return expr, nil
}

func (p *Parser) parsePrimary() (astpkg.Expression, error) {
	if p.Match(token.NUMBER) {
		tok := p.Previous()
		return &astpkg.NumberExpression{BaseNode: astpkg.BaseNode{NodeSpan: tok.Span}, Value: tok.Literal.(float64)}, nil
	}
	if p.Match(token.STRING) {
		tok := p.Previous()
		return &astpkg.StringExpression{BaseNode: astpkg.BaseNode{NodeSpan: tok.Span}, Value: tok.Literal.(string)}, nil
	}
	if p.Match(token.TRUE) {
		tok := p.Previous()
		return &astpkg.BoolExpression{BaseNode: astpkg.BaseNode{NodeSpan: tok.Span}, Value: true}, nil
	}
	if p.Match(token.FALSE) {
		tok := p.Previous()
		return &astpkg.BoolExpression{BaseNode: astpkg.BaseNode{NodeSpan: tok.Span}, Value: false}, nil
	}
	if p.Match(token.IDENTIFIER) {
		tok := p.Previous()
		return &astpkg.VariableExpression{BaseNode: astpkg.BaseNode{NodeSpan: tok.Span}, Name: tok.Lexeme}, nil
	}
	if p.Match(token.LEFT_PAREN) {
		open := p.Previous()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		close, err := p.Consume(token.RIGHT_PAREN, "ожидается ')' после выражения")
		if err != nil {
			return nil, err
		}
		return &astpkg.GroupingExpression{BaseNode: astpkg.BaseNode{NodeSpan: source.Merge(open.Span, close.Span)}, Expr: expr}, nil
	}
	if p.Match(token.LEFT_BRACKET) {
		open := p.Previous()
		elements, close, err := p.finishArguments(token.RIGHT_BRACKET, "ожидается ']' после элементов массива")
		if err != nil {
			return nil, err
		}
		return &astpkg.ArrayExpression{BaseNode: astpkg.BaseNode{NodeSpan: source.Merge(open.Span, close.Span)}, Elements: elements}, nil
	}
	return nil, p.errorAt(p.Peek(), "ожидается выражение")
}

func (p *Parser) finishArguments(end token.TokenType, endMessage string) ([]astpkg.Expression, token.Token, error) {
	args := make([]astpkg.Expression, 0)
	if !p.Check(end) {
		for {
			arg, err := p.parseExpression()
			if err != nil {
				return nil, token.Token{}, err
			}
			args = append(args, arg)
			if !p.Match(token.COMMA) {
				break
			}
		}
	}
	close, err := p.Consume(end, endMessage)
	return args, close, err
}

func (p *Parser) Match(types ...token.TokenType) bool {
	for _, tt := range types {
		if p.Check(tt) {
			p.Advance()
			return true
		}
	}
	return false
}
func (p *Parser) Check(tt token.TokenType) bool {
	if p.isAtEnd() {
		return tt == token.EOF
	}
	return p.Peek().Type == tt
}
func (p *Parser) Advance() token.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.Previous()
}
func (p *Parser) Peek() token.Token {
	if len(p.tokens) == 0 {
		return token.Token{Type: token.EOF}
	}
	if p.current >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.current]
}
func (p *Parser) Previous() token.Token {
	if len(p.tokens) == 0 {
		return token.Token{Type: token.EOF}
	}
	if p.current == 0 {
		return p.tokens[0]
	}
	return p.tokens[p.current-1]
}
func (p *Parser) Consume(tt token.TokenType, message string) (token.Token, error) {
	if p.Check(tt) {
		return p.Advance(), nil
	}
	return token.Token{}, p.errorAt(p.Peek(), message)
}
func (p *Parser) isAtEnd() bool { return p.Peek().Type == token.EOF }
func (p *Parser) errorAt(tok token.Token, message string) error {
	return &ParseError{Message: message, Span: tok.Span}
}
