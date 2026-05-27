package main

import (
	"bytes"
	"flag"
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
	file := flag.String("file", "./examples/final_demo.cl", "source file")
	showTokens := flag.Bool("tokens", false, "print tokens")
	showAST := flag.Bool("ast", false, "print AST")
	showOpt := flag.Bool("optimized", false, "print optimized AST")
	showIR := flag.Bool("llmr", true, "print LLMR intermediate representation")
	run := flag.Bool("run", true, "execute program with tree interpreter")
	noOpt := flag.Bool("no-opt", false, "disable AST optimizations before LLMR/interpreter")
	flag.Parse()

	data, err := os.ReadFile(*file)
	if err != nil {
		fmt.Println("read error:", err)
		os.Exit(1)
	}
	sourceText := string(data)

	l := lexer.New(sourceText)
	tokens, err := l.ScanTokens()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if *showTokens {
		fmt.Println("TOKENS:")
		for _, tok := range tokens {
			fmt.Println(tok)
		}
		fmt.Println()
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		fmt.Println("PARSER ERROR:")
		fmt.Println(err)
		os.Exit(1)
	}
	if *showAST {
		fmt.Println("AST:")
		fmt.Println(printer.PrintProgram(program))
		fmt.Println()
	}

	analyzer := semantic.New()
	semanticErrors := analyzer.Analyze(program)
	for _, warning := range analyzer.Warnings() {
		fmt.Println("WARNING:", warning)
	}
	if len(semanticErrors) > 0 {
		fmt.Println("SEMANTIC ERRORS:")
		for _, err := range semanticErrors {
			fmt.Println("-", err)
		}
		os.Exit(1)
	}
	fmt.Println("SEMANTIC: OK")

	finalProgram := program
	if !*noOpt {
		finalProgram = optimizer.New().Optimize(program)
	}
	if *showOpt {
		fmt.Println("OPTIMIZED AST:")
		fmt.Println(printer.PrintProgram(finalProgram))
		fmt.Println()
	}
	if *showIR {
		fmt.Println("LLMR:")
		fmt.Print(llmr.NewTranslator().Translate(finalProgram).String())
		fmt.Println()
	}
	if *run {
		var out bytes.Buffer
		if err := interpreter.New(&out).Interpret(finalProgram); err != nil {
			fmt.Println("RUNTIME ERROR:")
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("OUTPUT:")
		fmt.Print(out.String())
	}
}
