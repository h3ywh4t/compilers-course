package semantic

import (
	"fmt"

	astpkg "compilerlabs/ast"
	"compilerlabs/source"
	"compilerlabs/token"
	"compilerlabs/types"
)

type Analyzer struct {
	environment   *Environment
	errors        []error
	warnings      []error
	functionDepth int
	returnTypes   []types.Type
}

func New() *Analyzer { return &Analyzer{environment: NewEnvironment(nil)} }

func (a *Analyzer) Analyze(statements []astpkg.Statement) []error {
	for _, stmt := range statements {
		a.visitStatement(stmt)
	}
	a.checkUnusedVariables(a.environment)
	return append([]error(nil), a.errors...)
}
func (a *Analyzer) Warnings() []error { return append([]error(nil), a.warnings...) }
func (a *Analyzer) addError(span source.Span, msg string) {
	a.errors = append(a.errors, &Error{Message: msg, Span: span})
}
func (a *Analyzer) addWarning(span source.Span, msg string) {
	a.warnings = append(a.warnings, &Warning{Message: msg, Span: span})
}

func (a *Analyzer) visitStatement(statement astpkg.Statement) {
	switch stmt := statement.(type) {
	case *astpkg.VarStatement:
		a.analyzeVarStatement(stmt)
	case *astpkg.PrintStatement:
		a.visitExpression(stmt.Expr)
	case *astpkg.ExpressionStatement:
		a.visitExpression(stmt.Expr)
	case *astpkg.BlockStatement:
		a.analyzeBlockStatement(stmt)
	case *astpkg.IfStatement:
		a.analyzeIfStatement(stmt)
	case *astpkg.WhileStatement:
		a.analyzeWhileStatement(stmt)
	case *astpkg.FunctionStatement:
		a.analyzeFunctionStatement(stmt)
	case *astpkg.ReturnStatement:
		a.analyzeReturnStatement(stmt)
	default:
		a.addError(statement.Span(), fmt.Sprintf("Неподдерживаемая инструкция: %T", statement))
	}
}

func (a *Analyzer) analyzeVarStatement(stmt *astpkg.VarStatement) {
	declaredType := types.Unknown
	if stmt.TypeAnnotation != nil {
		declaredType = *stmt.TypeAnnotation
	}

	if !a.environment.DefineVariable(stmt.Name, false, declaredType) {
		a.addError(stmt.Span(), fmt.Sprintf("Переменная '%s' уже объявлена в этой области видимости.", stmt.Name))
		if stmt.Initializer != nil {
			a.visitExpression(stmt.Initializer)
		}
		return
	}

	if stmt.Initializer != nil {
		initType := a.visitExpression(stmt.Initializer)
		if stmt.TypeAnnotation != nil && !types.Compatible(declaredType, initType) {
			a.addError(stmt.Initializer.Span(), fmt.Sprintf("Ошибка типов: переменная '%s' объявлена как %s, но получает %s.", stmt.Name, declaredType, initType))
		}
		a.environment.SetInitialized(stmt.Name, initType)
	}
}

func (a *Analyzer) analyzeBlockStatement(stmt *astpkg.BlockStatement) {
	previous := a.environment
	a.environment = NewEnvironment(previous)
	for _, inner := range stmt.Statements {
		a.visitStatement(inner)
	}
	a.checkUnusedVariables(a.environment)
	a.environment = previous
}

func (a *Analyzer) analyzeIfStatement(stmt *astpkg.IfStatement) {
	condType := a.visitExpression(stmt.Condition)
	if condType.Kind != types.BoolKind && condType.Kind != types.UnknownKind {
		a.addError(stmt.Condition.Span(), fmt.Sprintf("Условие if должно иметь тип Bool, получено: %s.", condType))
	}
	a.visitStatement(stmt.ThenBranch)
	if stmt.ElseBranch != nil {
		a.visitStatement(stmt.ElseBranch)
	}
}

func (a *Analyzer) analyzeWhileStatement(stmt *astpkg.WhileStatement) {
	condType := a.visitExpression(stmt.Condition)
	if condType.Kind != types.BoolKind && condType.Kind != types.UnknownKind {
		a.addError(stmt.Condition.Span(), fmt.Sprintf("Условие while должно иметь тип Bool, получено: %s.", condType))
	}
	a.visitStatement(stmt.Body)
}

func (a *Analyzer) analyzeFunctionStatement(stmt *astpkg.FunctionStatement) {
	paramTypes := make([]types.Type, len(stmt.Parameters))
	for idx := range stmt.Parameters {
		paramTypes[idx] = types.Unknown
		if idx < len(stmt.ParameterTypes) && stmt.ParameterTypes[idx] != nil {
			paramTypes[idx] = *stmt.ParameterTypes[idx]
		}
	}
	returnType := types.Unknown
	if stmt.ReturnType != nil {
		returnType = *stmt.ReturnType
	}

	if !a.environment.DefineFunctionWithSignature(stmt.Name, paramTypes, returnType) {
		a.addError(stmt.Span(), fmt.Sprintf("Функция '%s' уже объявлена в этой области видимости.", stmt.Name))
		return
	}

	previous := a.environment
	a.environment = NewEnvironment(previous)
	a.functionDepth++
	a.returnTypes = append(a.returnTypes, returnType)

	seen := make(map[string]struct{})
	for idx, param := range stmt.Parameters {
		if _, ok := seen[param]; ok {
			a.addError(stmt.Span(), fmt.Sprintf("Параметр '%s' повторяется в функции '%s'.", param, stmt.Name))
			continue
		}
		seen[param] = struct{}{}
		a.environment.DefineVariable(param, true, paramTypes[idx])
	}
	for _, inner := range stmt.Body.Statements {
		a.visitStatement(inner)
	}
	a.checkUnusedVariables(a.environment)
	a.returnTypes = a.returnTypes[:len(a.returnTypes)-1]
	a.functionDepth--
	a.environment = previous
}

func (a *Analyzer) analyzeReturnStatement(stmt *astpkg.ReturnStatement) {
	if a.functionDepth == 0 {
		a.addError(stmt.Span(), "return можно использовать только внутри функции.")
	}
	valueType := types.Nil
	if stmt.Value != nil {
		valueType = a.visitExpression(stmt.Value)
	}
	if len(a.returnTypes) > 0 {
		expected := a.returnTypes[len(a.returnTypes)-1]
		if expected.Kind != types.UnknownKind && !types.Compatible(expected, valueType) {
			a.addError(stmt.Span(), fmt.Sprintf("Ошибка типов return: функция должна возвращать %s, но возвращает %s.", expected, valueType))
		}
	}
}

func (a *Analyzer) visitExpression(expression astpkg.Expression) types.Type {
	switch expr := expression.(type) {
	case *astpkg.NumberExpression:
		return types.Number
	case *astpkg.StringExpression:
		return types.String
	case *astpkg.BoolExpression:
		return types.Bool
	case *astpkg.VariableExpression:
		return a.analyzeVariableExpression(expr)
	case *astpkg.AssignExpression:
		return a.analyzeAssignExpression(expr)
	case *astpkg.BinaryExpression:
		return a.analyzeBinaryExpression(expr)
	case *astpkg.UnaryExpression:
		return a.analyzeUnaryExpression(expr)
	case *astpkg.GroupingExpression:
		return a.visitExpression(expr.Expr)
	case *astpkg.ArrayExpression:
		return a.analyzeArrayExpression(expr)
	case *astpkg.IndexExpression:
		return a.analyzeIndexExpression(expr)
	case *astpkg.IndexAssignExpression:
		return a.analyzeIndexAssignExpression(expr)
	case *astpkg.CallExpression:
		return a.analyzeCallExpression(expr)
	default:
		a.addError(expression.Span(), fmt.Sprintf("Неподдерживаемое выражение: %T", expression))
		return types.Unknown
	}
}

func (a *Analyzer) analyzeVariableExpression(expr *astpkg.VariableExpression) types.Type {
	symbol := a.environment.GetVariable(expr.Name)
	if symbol == nil {
		a.addError(expr.Span(), fmt.Sprintf("Использование необъявленной переменной '%s'.", expr.Name))
		return types.Unknown
	}
	symbol.IsUsed = true
	if !symbol.IsInitialized {
		a.addError(expr.Span(), fmt.Sprintf("Использование неинициализированной переменной '%s'.", expr.Name))
	}
	return symbol.Type
}

func (a *Analyzer) analyzeAssignExpression(expr *astpkg.AssignExpression) types.Type {
	valueType := a.visitExpression(expr.Value)
	symbol := a.environment.GetVariable(expr.Name)
	if symbol == nil {
		a.addError(expr.Span(), fmt.Sprintf("Попытка записи в необъявленную переменную '%s'.", expr.Name))
		return valueType
	}
	if symbol.Type.Kind == types.FunctionKind {
		a.addError(expr.Span(), fmt.Sprintf("Нельзя присвоить значение функции '%s'.", expr.Name))
		return symbol.Type
	}
	if !types.Compatible(symbol.Type, valueType) {
		a.addError(expr.Span(), fmt.Sprintf("Ошибка типов: нельзя присвоить %s переменной '%s' типа %s.", valueType, expr.Name, symbol.Type))
	}
	symbol.IsInitialized = true
	if symbol.Type.Kind == types.UnknownKind && valueType.Kind != types.UnknownKind {
		symbol.Type = valueType
	}
	return symbol.Type
}

func (a *Analyzer) analyzeArrayExpression(expr *astpkg.ArrayExpression) types.Type {
	if len(expr.Elements) == 0 {
		return types.ArrayOf(types.Unknown)
	}
	first := a.visitExpression(expr.Elements[0])
	for _, element := range expr.Elements[1:] {
		t := a.visitExpression(element)
		if !types.Compatible(first, t) || !types.Compatible(t, first) {
			a.addError(element.Span(), fmt.Sprintf("Все элементы массива должны быть одного типа: ожидался %s, получено %s.", first, t))
		}
		if first.Kind == types.UnknownKind && t.Kind != types.UnknownKind {
			first = t
		}
	}
	return types.ArrayOf(first)
}

func (a *Analyzer) analyzeIndexExpression(expr *astpkg.IndexExpression) types.Type {
	arrayType := a.visitExpression(expr.Array)
	indexType := a.visitExpression(expr.Index)
	if indexType.Kind != types.NumberKind && indexType.Kind != types.UnknownKind {
		a.addError(expr.Index.Span(), fmt.Sprintf("Индекс массива должен иметь тип Number, получено: %s.", indexType))
	}
	if arrayType.Kind != types.ArrayKind {
		if arrayType.Kind != types.UnknownKind {
			a.addError(expr.Array.Span(), fmt.Sprintf("Индексация применима только к массивам, получено: %s.", arrayType))
		}
		return types.Unknown
	}
	if arrayType.Element == nil {
		return types.Unknown
	}
	return *arrayType.Element
}

func (a *Analyzer) analyzeIndexAssignExpression(expr *astpkg.IndexAssignExpression) types.Type {
	arrayType := a.visitExpression(expr.Array)
	indexType := a.visitExpression(expr.Index)
	valueType := a.visitExpression(expr.Value)
	if indexType.Kind != types.NumberKind && indexType.Kind != types.UnknownKind {
		a.addError(expr.Index.Span(), fmt.Sprintf("Индекс массива должен иметь тип Number, получено: %s.", indexType))
	}
	if arrayType.Kind != types.ArrayKind {
		if arrayType.Kind != types.UnknownKind {
			a.addError(expr.Array.Span(), fmt.Sprintf("Присваивание по индексу применимо только к массивам, получено: %s.", arrayType))
		}
		return types.Unknown
	}
	if arrayType.Element != nil && !types.Compatible(*arrayType.Element, valueType) {
		a.addError(expr.Value.Span(), fmt.Sprintf("Нельзя положить %s в массив элементов %s.", valueType, arrayType.Element.String()))
	}
	if arrayType.Element == nil {
		return types.Unknown
	}
	return *arrayType.Element
}

func (a *Analyzer) analyzeCallExpression(expr *astpkg.CallExpression) types.Type {
	symbol := a.environment.GetVariable(expr.CalleeName)
	if symbol == nil || symbol.Type.Kind != types.FunctionKind {
		a.addError(expr.Span(), fmt.Sprintf("Вызов неопределенной функции '%s'.", expr.CalleeName))
		return types.Unknown
	}
	symbol.IsUsed = true
	if len(expr.Arguments) != symbol.Arity {
		a.addError(expr.Span(), fmt.Sprintf("Неверное количество аргументов при вызове '%s': ожидалось %d, получено %d.", expr.CalleeName, symbol.Arity, len(expr.Arguments)))
	}
	for idx, arg := range expr.Arguments {
		argType := a.visitExpression(arg)
		if idx < len(symbol.ParamTypes) && !types.Compatible(symbol.ParamTypes[idx], argType) {
			a.addError(arg.Span(), fmt.Sprintf("Ошибка типов аргумента %d при вызове '%s': ожидался %s, получено %s.", idx+1, expr.CalleeName, symbol.ParamTypes[idx], argType))
		}
	}
	if symbol.Type.Return != nil {
		return *symbol.Type.Return
	}
	return symbol.ReturnType
}

func (a *Analyzer) analyzeBinaryExpression(expr *astpkg.BinaryExpression) types.Type {
	left := a.visitExpression(expr.Left)
	right := a.visitExpression(expr.Right)
	if left.Kind == types.UnknownKind || right.Kind == types.UnknownKind {
		return types.Unknown
	}
	switch expr.Operator.Type {
	case token.PLUS:
		if left.Kind == types.NumberKind && right.Kind == types.NumberKind {
			return types.Number
		}
		if left.Kind == types.StringKind || right.Kind == types.StringKind {
			return types.String
		}
		a.addError(expr.Span(), fmt.Sprintf("Оператор '+' нельзя применить к %s и %s.", left, right))
		return types.Unknown
	case token.MINUS, token.STAR, token.SLASH:
		if left.Kind == types.NumberKind && right.Kind == types.NumberKind {
			return types.Number
		}
		a.addError(expr.Span(), fmt.Sprintf("Оператор '%s' работает только с Number.", expr.Operator.Lexeme))
		return types.Unknown
	case token.LESS, token.LESS_EQUAL, token.GREATER, token.GREATER_EQUAL:
		if left.Kind == types.NumberKind && right.Kind == types.NumberKind {
			return types.Bool
		}
		a.addError(expr.Span(), "Операторы сравнения работают только с Number.")
		return types.Unknown
	case token.EQUAL_EQUAL, token.BANG_EQUAL:
		if !types.Equal(left, right) {
			a.addWarning(expr.Span(), fmt.Sprintf("Сравнение разных типов: %s и %s.", left, right))
		}
		return types.Bool
	case token.AND_AND, token.OR_OR:
		if left.Kind == types.BoolKind && right.Kind == types.BoolKind {
			return types.Bool
		}
		a.addError(expr.Span(), "Логические операторы требуют Bool.")
		return types.Unknown
	default:
		return types.Unknown
	}
}

func (a *Analyzer) analyzeUnaryExpression(expr *astpkg.UnaryExpression) types.Type {
	right := a.visitExpression(expr.Right)
	if right.Kind == types.UnknownKind {
		return types.Unknown
	}
	switch expr.Operator.Type {
	case token.MINUS:
		if right.Kind != types.NumberKind {
			a.addError(expr.Span(), "Унарный минус применяется только к Number.")
			return types.Unknown
		}
		return types.Number
	case token.BANG:
		if right.Kind != types.BoolKind {
			a.addError(expr.Span(), "Оператор ! применяется только к Bool.")
			return types.Unknown
		}
		return types.Bool
	default:
		return right
	}
}

func (a *Analyzer) checkUnusedVariables(env *Environment) {
	for _, symbol := range env.GetLocalVariables() {
		if !symbol.IsUsed && symbol.Type.Kind != types.FunctionKind {
			a.addWarning(source.Span{}, fmt.Sprintf("Переменная '%s' объявлена, но не используется.", symbol.Name))
		}
	}
}
