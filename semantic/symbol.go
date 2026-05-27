package semantic

import "compilerlabs/types"

type SymbolInfo struct {
	Name          string
	Type          types.Type
	IsInitialized bool
	IsUsed        bool
	Arity         int
	ParamTypes    []types.Type
	ReturnType    types.Type
}
