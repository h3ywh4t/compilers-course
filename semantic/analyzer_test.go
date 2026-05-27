package semantic

import (
	"compilerlabs/lexer"
	"compilerlabs/parser"
	"testing"
)

func analyzeSource(t *testing.T, src string) []error {
	t.Helper()
	tokens, err := lexer.New(src).ScanTokens()
	if err != nil {
		t.Fatalf("lexer: %v", err)
	}
	program, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatalf("parser: %v", err)
	}
	return New().Analyze(program)
}

func TestSemanticOK(t *testing.T) {
	errs := analyzeSource(t, `fun f(x) { return x * 2; } var a = [1,2,3]; a[1] = f(10); print a[1];`)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}
func TestSemanticUndeclared(t *testing.T) {
	if len(analyzeSource(t, `print x;`)) == 0 {
		t.Fatal("expected error")
	}
}
func TestSemanticUninitialized(t *testing.T) {
	if len(analyzeSource(t, `var x; print x;`)) == 0 {
		t.Fatal("expected error")
	}
}
func TestSemanticTypeMismatch(t *testing.T) {
	if len(analyzeSource(t, `var x = 1; x = "s";`)) == 0 {
		t.Fatal("expected error")
	}
}
func TestSemanticArrayMixed(t *testing.T) {
	if len(analyzeSource(t, `var a = [1, "s"];`)) == 0 {
		t.Fatal("expected error")
	}
}

func TestSemantic_ExplicitVariableTypeOK(t *testing.T) {
	errs := analyzeSource(t, `var x : number = 10; print x;`)
	if len(errs) != 0 {
		t.Fatalf("expected 0 semantic errors, got %d: %v", len(errs), errs)
	}
}

func TestSemantic_ExplicitVariableTypeMismatch(t *testing.T) {
	errs := analyzeSource(t, `var x : number = "bad";`)
	if len(errs) != 1 {
		t.Fatalf("expected 1 semantic error, got %d: %v", len(errs), errs)
	}
}

func TestSemantic_FunctionArgumentAndReturnTypes(t *testing.T) {
	errs := analyzeSource(t, `
fun add(a : number, b : number) : number {
	return a + b;
}
print add(1, 2);
`)
	if len(errs) != 0 {
		t.Fatalf("expected 0 semantic errors, got %d: %v", len(errs), errs)
	}
}

func TestSemantic_FunctionArgumentTypeMismatch(t *testing.T) {
	errs := analyzeSource(t, `
fun add(a : number, b : number) : number {
	return a + b;
}
print add(1, "x");
`)
	if len(errs) == 0 {
		t.Fatal("expected semantic errors, got nil")
	}
}

func TestSemantic_FunctionReturnTypeMismatch(t *testing.T) {
	errs := analyzeSource(t, `
fun bad() : number {
	return "oops";
}
`)
	if len(errs) != 1 {
		t.Fatalf("expected 1 semantic error, got %d: %v", len(errs), errs)
	}
}
