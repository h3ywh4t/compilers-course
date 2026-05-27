package source

import "fmt"

type Position struct {
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

type Span struct {
	Start Position
	End   Position
}

func (s Span) String() string {
	return fmt.Sprintf("%s-%s", s.Start, s.End)
}

func Merge(a, b Span) Span {
	return Span{Start: a.Start, End: b.End}
}
