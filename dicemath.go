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

func (o Operator) Eval(l, r int) int {
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
		return die_roll(int(l), int(r))
	}
	panic("unsupported operator")
}

func (v *Value) Eval(ctx Context) int {
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

func (f *Factor) Eval(ctx Context) int {
	b := f.Base.Eval(ctx)
	if f.Exponent != nil {
		return int(math.Pow(float64(b), float64(f.Exponent.Eval(ctx))))
	}
	return b
}

func (t *Term) Eval(ctx Context) int {
	n := 0
	if t.Left != nil {
		n = t.Left.Eval(ctx)
		if t.Right != nil {
			for _, r := range t.Right {
				n = r.Operator.Eval(n, r.Factor.Eval(ctx))
			}
		}
	}
	return n
}

func (e *Expression) Eval(ctx Context) int {
	l := 0
	if e.Left != nil {
		l = e.Left.Eval(ctx)
		if e.Right != nil {
			for _, r := range e.Right {
				l = r.Operator.Eval(l, r.Term.Eval(ctx))
			}
		}
	}
	return l
}

type Context map[string]int

func Parse(calc string) string {
	rollLexer := lexer.MustSimple([]lexer.Rule{
		{"Comment", `(?:;)[^,]*`, nil},
		{"Ident", `[a-zA-Z]+`, nil},
		{"Number", `(?:\d*)?\d+`, nil},
		{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`, nil},
		{"Whitespace", `[ \t\n\r]+`, nil},
	})

	var parser = participle.MustBuild(
		&Expression{},
		participle.Lexer(rollLexer),
		participle.Elide("Comment", "Whitespace"),
	)

	fmt.Println(parser)

	expr := &Expression{}
	parser.ParseString("", calc, expr)

	ctx := make(Context)

	return fmt.Sprintf("%s = %d", expr, expr.Eval(ctx))
}
