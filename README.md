# Compilers Course Project

<div align="center">

**A small compiler project written in Go**

Lexer · Parser · AST · Semantic Analysis · Interpreter · Optimizer · LLMR

</div>

---

## About

This is an educational compiler for a small typed language.  
The project shows the main compiler stages: from source code to tokens, AST, semantic checks, optimization, execution, and intermediate representation.

```text
source code → tokens → AST → semantic analysis → optimized AST → interpreter / LLMR
```

---

## Features

- lexical analysis
- recursive descent parser
- AST generation
- semantic analysis
- simple type system
- interpreter with runtime environment
- arrays and functions
- constant folding
- dead code elimination
- AST to LLMR translation
- readable error messages with source positions

---

## Language

The language supports:

- `number`, `string`, and `bool`
- variables
- arrays
- `if / else`
- `while`
- functions
- `return`
- `print`

Example:

```text
var title : string = "compiler";
var values : number[] = [1, 2, 3];

fun double(x : number) : number {
    return x * 2;
}

values[1] = double(10);
print title + ": " + values[1];
```

---

## Project structure

```text
ast/           AST nodes
lexer/         lexical analyzer
parser/        parser
semantic/      semantic checks and type analysis
runtime/       runtime environment
interpreter/   program execution
optimizer/     AST optimizations
llmr/          intermediate representation
examples/      sample programs
cmd/           CLI entry points
```

---

## Run

Run tests:

```bash
go test ./...
```

Run a program:

```bash
go run ./cmd/compilerlab --file ./examples/final_demo.cl
```

Show compiler stages:

```bash
go run ./cmd/compilerlab --file ./examples/ok_typed.cl --tokens --ast --optimized
```

---

## Compiler stages

| Stage | Description |
|---|---|
| Lexer | Splits source code into tokens |
| Parser | Builds an AST |
| Semantic analyzer | Checks variables, scopes, and types |
| Optimizer | Simplifies AST and removes dead code |
| Interpreter | Executes the program |
| LLMR translator | Converts AST to intermediate representation |