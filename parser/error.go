package parser

import (
	"compilerlabs/source"
	"fmt"
)

type ParseError struct {
	Message string
	Span    source.Span
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parser error at %s: %s", e.Span.Start, e.Message)
}
