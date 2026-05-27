package ast

import "compilerlabs/source"

type Node interface{ Span() source.Span }

type Expression interface {
	Node
	expressionNode()
}

type Statement interface {
	Node
	statementNode()
}

type BaseNode struct{ NodeSpan source.Span }

func (n BaseNode) Span() source.Span { return n.NodeSpan }
