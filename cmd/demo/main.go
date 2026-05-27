package main

import (
	"bytes"
	"fmt"
	"os"

	"compilerlabs/interpreter"
	"compilerlabs/lexer"
	"compilerlabs/llmr"
	"compilerlabs/optimizer"
	"compilerlabs/parser"
	"compilerlabs/printer"
	"compilerlabs/semantic"
)

func main() {
	path := "./examples/final_demo.cl"
	if len(os.Args) >= 2 {
		path = os.Args[1]
	}
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("read error:", err)
		os.Exit(1)
	}

	l := lexer.New(string(data))
	tokens, err := l.ScanTokens()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("TOKENS:")
	for _, tok := range tokens {
		fmt.Println(tok)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		fmt.Println("PARSER ERROR:")
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("\nAST:")
	fmt.Println(printer.PrintProgram(program))

	analyzer := semantic.New()
	errs := analyzer.Analyze(program)
	for _, warning := range analyzer.Warnings() {
		fmt.Println("WARNING:", warning)
	}
	if len(errs) > 0 {
		fmt.Println("SEMANTIC ERRORS:")
		for _, err := range errs {
			fmt.Println("-", err)
		}
		os.Exit(1)
	}
	fmt.Println("SEMANTIC: OK")

	optimized := optimizer.New().Optimize(program)
	fmt.Println("\nOPTIMIZED AST:")
	fmt.Println(printer.PrintProgram(optimized))
	fmt.Println("\nLLMR:")
	fmt.Print(llmr.NewTranslator().Translate(optimized).String())
	var out bytes.Buffer
	if err := interpreter.New(&out).Interpret(optimized); err != nil {
		fmt.Println("RUNTIME ERROR:")
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("\nOUTPUT:")
	fmt.Print(out.String())
}
