package semantic

import "compilerlabs/types"

type Environment struct {
	parent    *Environment
	variables map[string]*SymbolInfo
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{parent: parent, variables: make(map[string]*SymbolInfo)}
}

func (e *Environment) DefineVariable(name string, isInitialized bool, typ types.Type) bool {
	if _, exists := e.variables[name]; exists {
		return false
	}
	e.variables[name] = &SymbolInfo{Name: name, Type: typ, IsInitialized: isInitialized}
	return true
}

func (e *Environment) DefineFunction(name string, arity int) bool {
	paramTypes := make([]types.Type, arity)
	for i := range paramTypes {
		paramTypes[i] = types.Unknown
	}
	return e.DefineFunctionWithSignature(name, paramTypes, types.Unknown)
}

func (e *Environment) DefineFunctionWithSignature(name string, paramTypes []types.Type, returnType types.Type) bool {
	if _, exists := e.variables[name]; exists {
		return false
	}
	arity := len(paramTypes)
	e.variables[name] = &SymbolInfo{
		Name:          name,
		Type:          types.FunctionWithSignature(paramTypes, returnType),
		IsInitialized: true,
		Arity:         arity,
		ParamTypes:    append([]types.Type(nil), paramTypes...),
		ReturnType:    returnType,
	}
	return true
}

func (e *Environment) GetLocalVariables() []*SymbolInfo {
	out := make([]*SymbolInfo, 0, len(e.variables))
	for _, symbol := range e.variables {
		out = append(out, symbol)
	}
	return out
}

func (e *Environment) IsVariableDefined(name string) bool { return e.GetVariable(name) != nil }
func (e *Environment) GetVariable(name string) *SymbolInfo {
	if symbol, ok := e.variables[name]; ok {
		return symbol
	}
	if e.parent != nil {
		return e.parent.GetVariable(name)
	}
	return nil
}
func (e *Environment) SetInitialized(name string, typ types.Type) {
	s := e.GetVariable(name)
	if s != nil {
		s.IsInitialized = true
		if s.Type.Kind == types.UnknownKind && typ.Kind != types.UnknownKind {
			s.Type = typ
		}
	}
}
