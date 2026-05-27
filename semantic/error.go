package semantic

import (
	"compilerlabs/source"
	"fmt"
)

type Error struct {
	Message string
	Span    source.Span
}

func (e *Error) Error() string {
	return fmt.Sprintf("semantic error at %s: %s", e.Span.Start, e.Message)
}

type Warning struct {
	Message string
	Span    source.Span
}

func (w *Warning) Error() string {
	return fmt.Sprintf("semantic warning at %s: %s", w.Span.Start, w.Message)
}
