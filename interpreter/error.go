package interpreter

import (
	"compilerlabs/source"
	"fmt"
)

type RuntimeError struct {
	Message string
	Span    source.Span
}

func (e *RuntimeError) Error() string {
	return fmt.Sprintf("runtime error at %s: %s", e.Span.Start, e.Message)
}

type returnSignal struct{ value any }

func (r *returnSignal) Error() string { return "return" }
