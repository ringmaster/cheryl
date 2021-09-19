package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
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
	Number        *int        `  @Number`
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
	Left    *Term     `@@`
	Right   []*OpTerm `@@*`
	Comment *string   `(@Comment)?`
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
		return fmt.Sprintf("%d", *v.Number)
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
	if e.Comment != nil {
		out = append(out, *e.Comment)
	}
	return strings.Join(out, " ")
}

// Evaluation

func die_roll(ctx Parser, count int, sides int) int {
	result := 0

	//fmt.Printf("Rolling %dd%d:\n", count, sides)
	for i := 0; i < count; i++ {
		die_result := ctx.randomizer(1, sides)
		//fmt.Printf("  Rolling d%d: %d\n", sides, die_result)
		result += die_result
	}

	return result
}

func (o Operator) Eval(ctx Parser, l, r int) int {
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
		return die_roll(ctx, int(l), int(r))
	}
	panic("unsupported operator")
}

func (v *Value) Eval(ctx Parser) int {
	switch {
	case v.Number != nil:
		return *v.Number
	case v.Variable != nil:
		value, ok := ctx.variables[*v.Variable]
		if !ok {
			panic("no such variable " + *v.Variable)
		}
		return value
	default:
		return v.Subexpression.Eval(ctx)
	}
}

func (f *Factor) Eval(ctx Parser) int {
	b := f.Base.Eval(ctx)
	if f.Exponent != nil {
		return int(math.Pow(float64(b), float64(f.Exponent.Eval(ctx))))
	}
	return b
}

func (t *Term) Eval(ctx Parser) int {
	n := 0
	if t.Left != nil {
		n = t.Left.Eval(ctx)
		if t.Right != nil {
			for _, r := range t.Right {
				n = r.Operator.Eval(ctx, n, r.Factor.Eval(ctx))
			}
		}
	}
	return n
}

func (e *Expression) Eval(ctx Parser) int {
	l := 0
	if e.Left != nil {
		l = e.Left.Eval(ctx)
		if e.Right != nil {
			for _, r := range e.Right {
				l = r.Operator.Eval(ctx, l, r.Term.Eval(ctx))
			}
		}
	}
	return l
}

type Parser struct {
	variables  map[string]int
	randomizer func(min int, max int) int
	lexer      *lexer.StatefulDefinition
}

func Parse(calc string, opts ...func(*Parser)) string {

	p := &Parser{}

	p.randomizer = func(min int, max int) int {
		return rand.Intn(max-min) + min
	}

	p.lexer = lexer.MustSimple([]lexer.Rule{
		{"Comment", `(?:;)[^,]*`, nil},
		{"Ident", `[a-zA-Z]+`, nil},
		{"Number", `(?:\d*)?\d+`, nil},
		{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`, nil},
		{"Whitespace", `[ \t\n\r]+`, nil},
	})

	p.variables = make(map[string]int)

	p.randomizer = func(min int, max int) int { return rand.Intn(max-min) + min }

	// call option functions on instance to set options on it
	for _, opt := range opts {
		opt(p)
	}

	var parser = participle.MustBuild(
		&Expression{},
		participle.Lexer(p.lexer),
		participle.Elide("Comment", "Whitespace"),
	)

	/*
		// This outputs the parser as a human-readable rule set
		fmt.Println(parser)
	*/

	/*
		// This chunk of code outputs the generated lexer as Go code, which is 10x faster
		f, err := os.OpenFile("./rolllexer.go", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		lexer.ExperimentalGenerateLexer(f, "main", rollLexer)
	*/

	expr := &Expression{}
	parser.ParseString("", calc, expr)

	return fmt.Sprintf("%s = %d", expr, expr.Eval(*p))
}
