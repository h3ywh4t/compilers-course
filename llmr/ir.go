package llmr

import (
	"fmt"
	"strings"
)

type Instruction struct {
	Op     string
	Result string
	Args   []string
}

type Program struct{ Instructions []Instruction }

func (p Program) String() string {
	var b strings.Builder
	for _, ins := range p.Instructions {
		b.WriteString(ins.String())
		b.WriteByte('\n')
	}
	return b.String()
}

func (i Instruction) String() string {
	args := strings.Join(i.Args, ", ")
	if i.Result != "" {
		return fmt.Sprintf("%-8s = %-14s %s", i.Result, i.Op, args)
	}
	if args == "" {
		return i.Op
	}
	return fmt.Sprintf("%-14s %s", i.Op, args)
}
