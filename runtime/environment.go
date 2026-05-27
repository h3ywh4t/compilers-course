package runtime

import (
	astpkg "compilerlabs/ast"
	"fmt"
)

type FunctionValue struct {
	Declaration *astpkg.FunctionStatement
	Closure     *Environment
}

type Environment struct {
	parent    *Environment
	values    map[string]any
	functions map[string]*FunctionValue
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{parent: parent, values: make(map[string]any), functions: make(map[string]*FunctionValue)}
}

func (e *Environment) Define(name string, value any) { e.values[name] = value }
func (e *Environment) Assign(name string, value any) error {
	if _, ok := e.values[name]; ok {
		e.values[name] = value
		return nil
	}
	if e.parent != nil {
		return e.parent.Assign(name, value)
	}
	return fmt.Errorf("runtime error: неизвестная переменная '%s'", name)
}
func (e *Environment) Get(name string) (any, error) {
	if value, ok := e.values[name]; ok {
		return value, nil
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, fmt.Errorf("runtime error: неизвестная переменная '%s'", name)
}
func (e *Environment) DefineFunction(name string, fn *FunctionValue) { e.functions[name] = fn }
func (e *Environment) GetFunction(name string) (*FunctionValue, error) {
	if fn, ok := e.functions[name]; ok {
		return fn, nil
	}
	if e.parent != nil {
		return e.parent.GetFunction(name)
	}
	return nil, fmt.Errorf("runtime error: неизвестная функция '%s'", name)
}
