package types

import "fmt"

type Kind string

const (
	UnknownKind  Kind = "Unknown"
	NumberKind   Kind = "Number"
	StringKind   Kind = "String"
	BoolKind     Kind = "Bool"
	ArrayKind    Kind = "Array"
	FunctionKind Kind = "Function"
	NilKind      Kind = "Nil"
)

type Type struct {
	Kind       Kind
	Element    *Type
	Arity      int
	ParamTypes []Type
	Return     *Type
}

var (
	Unknown = Type{Kind: UnknownKind}
	Number  = Type{Kind: NumberKind}
	String  = Type{Kind: StringKind}
	Bool    = Type{Kind: BoolKind}
	Nil     = Type{Kind: NilKind}
)

func ArrayOf(element Type) Type { return Type{Kind: ArrayKind, Element: &element} }
func Function(arity int) Type   { return Type{Kind: FunctionKind, Arity: arity} }
func FunctionWithSignature(params []Type, returnType Type) Type {
	copied := append([]Type(nil), params...)
	return Type{Kind: FunctionKind, Arity: len(params), ParamTypes: copied, Return: &returnType}
}

func (t Type) String() string {
	switch t.Kind {
	case ArrayKind:
		if t.Element == nil {
			return "Array<Unknown>"
		}
		return fmt.Sprintf("Array<%s>", t.Element.String())
	case FunctionKind:
		if len(t.ParamTypes) == 0 && t.Return == nil {
			return fmt.Sprintf("Function/%d", t.Arity)
		}
		parts := make([]string, len(t.ParamTypes))
		for i, p := range t.ParamTypes {
			parts[i] = p.String()
		}
		ret := Unknown
		if t.Return != nil {
			ret = *t.Return
		}
		return fmt.Sprintf("Function(%v) -> %s", parts, ret.String())
	default:
		return string(t.Kind)
	}
}

func Equal(a, b Type) bool {
	if a.Kind != b.Kind {
		return false
	}
	if a.Kind == ArrayKind {
		if a.Element == nil || b.Element == nil {
			return true
		}
		return Equal(*a.Element, *b.Element)
	}
	if a.Kind == FunctionKind {
		return a.Arity == b.Arity
	}
	return true
}

func Compatible(expected, actual Type) bool {
	if expected.Kind == UnknownKind || actual.Kind == UnknownKind {
		return true
	}
	return Equal(expected, actual)
}
