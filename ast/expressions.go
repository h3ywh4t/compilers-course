package ast

import "compilerlabs/token"

type NumberExpression struct {
	BaseNode
	Value float64
}

func (*NumberExpression) expressionNode() {}

type StringExpression struct {
	BaseNode
	Value string
}

func (*StringExpression) expressionNode() {}

type BoolExpression struct {
	BaseNode
	Value bool
}

func (*BoolExpression) expressionNode() {}

type VariableExpression struct {
	BaseNode
	Name string
}

func (*VariableExpression) expressionNode() {}

type GroupingExpression struct {
	BaseNode
	Expr Expression
}

func (*GroupingExpression) expressionNode() {}

type UnaryExpression struct {
	BaseNode
	Operator token.Token
	Right    Expression
}

func (*UnaryExpression) expressionNode() {}

type BinaryExpression struct {
	BaseNode
	Left     Expression
	Operator token.Token
	Right    Expression
}

func (*BinaryExpression) expressionNode() {}

type AssignExpression struct {
	BaseNode
	Name  string
	Value Expression
}

func (*AssignExpression) expressionNode() {}

type ArrayExpression struct {
	BaseNode
	Elements []Expression
}

func (*ArrayExpression) expressionNode() {}

type IndexExpression struct {
	BaseNode
	Array Expression
	Index Expression
}

func (*IndexExpression) expressionNode() {}

type IndexAssignExpression struct {
	BaseNode
	Array Expression
	Index Expression
	Value Expression
}

func (*IndexAssignExpression) expressionNode() {}

type CallExpression struct {
	BaseNode
	CalleeName string
	Arguments  []Expression
}

func (*CallExpression) expressionNode() {}
