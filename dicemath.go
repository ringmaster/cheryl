package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/alecthomas/participle/v2"
)

type Operator int

const (
	OpMul Operator = iota
	OpDiv
	OpAdd
	OpSub
	OpDice
)

var operatorMap = map[string]Operator{"+": OpAdd, "-": OpSub, "*": OpMul, "/": OpDiv, "d": OpDice}

func (o *Operator) Capture(s []string) error {
	*o = operatorMap[s[0]]
	return nil
}

// E --> T {( "+" | "-" ) T}
// T --> F {( "*" | "/" ) F}
// F --> B ["^" F]
// B --> v | "(" E ")" | "-" T

type Value struct {
	Number        *float64    `  @(Float|Int)`
	Variable      *string     `| @Ident`
	Subexpression *Expression `| "(" @@ ")"`
}

type Factor struct {
	Base     *Value `@@`
	Exponent *Value `( "^" @@ )?`
}

type OpFactor struct {
	Operator Operator `@("*" | "/")`
	Factor   *Factor  `@@`
}

type Term struct {
	Left  *Factor     `@@`
	Right []*OpFactor `@@*`
}

type OpTerm struct {
	Operator Operator `@("+" | "-" | "d")`
	Term     *Term    `@@`
}

type Expression struct {
	Left  *Term     `@@`
	Right []*OpTerm `@@*`
}

// Display

func (o Operator) String() string {
	switch o {
	case OpMul:
		return "*"
	case OpDiv:
		return "/"
	case OpSub:
		return "-"
	case OpAdd:
		return "+"
	case OpDice:
		return "d"
	}
	panic("unsupported operator")
}

func (v *Value) String() string {
	if v.Number != nil {
		return fmt.Sprintf("%g", *v.Number)
	}
	if v.Variable != nil {
		return *v.Variable
	}
	return "(" + v.Subexpression.String() + ")"
}

func (f *Factor) String() string {
	out := f.Base.String()
	if f.Exponent != nil {
		out += " ^ " + f.Exponent.String()
	}
	return out
}

func (o *OpFactor) String() string {
	return fmt.Sprintf("%s %s", o.Operator, o.Factor)
}

func (t *Term) String() string {
	out := []string{t.Left.String()}
	for _, r := range t.Right {
		out = append(out, r.String())
	}
	return strings.Join(out, " ")
}

func (o *OpTerm) String() string {
	return fmt.Sprintf("%s %s", o.Operator, o.Term)
}

func (e *Expression) String() string {
	out := []string{e.Left.String()}
	for _, r := range e.Right {
		out = append(out, r.String())
	}
	return strings.Join(out, " ")
}

// Evaluation

func die_roll(count int, sides int) int {
	result := 0

	fmt.Printf("Rolling %dd%d:\n", count, sides)
	for i := 0; i < count; i++ {
		die_result := rand.Intn(sides) + 1
		fmt.Printf("  Rolling d%d: %d\n", sides, die_result)
		result += die_result
	}

	return result
}

func (o Operator) Eval(l, r float64) float64 {
	switch o {
	case OpMul:
		return l * r
	case OpDiv:
		return l / r
	case OpAdd:
		return l + r
	case OpSub:
		return l - r
	case OpDice:
		return float64(die_roll(int(l), int(r)))
	}
	panic("unsupported operator")
}

func (v *Value) Eval(ctx Context) float64 {
	switch {
	case v.Number != nil:
		return *v.Number
	case v.Variable != nil:
		value, ok := ctx[*v.Variable]
		if !ok {
			panic("no such variable " + *v.Variable)
		}
		return value
	default:
		return v.Subexpression.Eval(ctx)
	}
}

func (f *Factor) Eval(ctx Context) float64 {
	b := f.Base.Eval(ctx)
	if f.Exponent != nil {
		return math.Pow(b, f.Exponent.Eval(ctx))
	}
	return b
}

func (t *Term) Eval(ctx Context) float64 {
	n := t.Left.Eval(ctx)
	for _, r := range t.Right {
		n = r.Operator.Eval(n, r.Factor.Eval(ctx))
	}
	return n
}

func (e *Expression) Eval(ctx Context) int {
	l := e.Left.Eval(ctx)
	for _, r := range e.Right {
		l = r.Operator.Eval(l, r.Term.Eval(ctx))
	}
	return int(l)
}

type Context map[string]float64

func Parse(calc string) string {
	var parser = participle.MustBuild(&Expression{})

	fmt.Println(parser)

	expr := &Expression{}
	parser.ParseString("", calc, expr)

	ctx := make(Context)

	return fmt.Sprintf("%s = %f", expr, expr.Eval(ctx))
}
