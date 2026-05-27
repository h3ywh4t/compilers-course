package ast

import "compilerlabs/types"

type ExpressionStatement struct {
	BaseNode
	Expr Expression
}

func (*ExpressionStatement) statementNode() {}

type PrintStatement struct {
	BaseNode
	Expr Expression
}

func (*PrintStatement) statementNode() {}

type VarStatement struct {
	BaseNode
	Name           string
	TypeAnnotation *types.Type
	Initializer    Expression
}

func (*VarStatement) statementNode() {}

type BlockStatement struct {
	BaseNode
	Statements []Statement
}

func (*BlockStatement) statementNode() {}

type IfStatement struct {
	BaseNode
	Condition  Expression
	ThenBranch Statement
	ElseBranch Statement
}

func (*IfStatement) statementNode() {}

type WhileStatement struct {
	BaseNode
	Condition Expression
	Body      Statement
}

func (*WhileStatement) statementNode() {}

type FunctionStatement struct {
	BaseNode
	Name           string
	Parameters     []string
	ParameterTypes []*types.Type
	ReturnType     *types.Type
	Body           *BlockStatement
}

func (*FunctionStatement) statementNode() {}

type ReturnStatement struct {
	BaseNode
	Value Expression
}

func (*ReturnStatement) statementNode() {}
