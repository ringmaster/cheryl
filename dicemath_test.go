package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func harness(p *Parser) {
	p.randomizer = func(min int, max int) int { return 4 }
	p.variables["juice"] = 4
}

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestBasicMath(t *testing.T) {
	input := "3+4"
	output := Parse(input, harness)
	assert.Equal(t, "3 + 4 = 7", output, "formats and computes addition")

	input = "( ( 3+3*      12 -   30) /  3 ) ^2"
	output = Parse(input, harness)
	assert.Equal(t, "((3 + 3 * 12 - 30) / 3) ^ 2 = 9", output, "formats and computes arithmetic")
}

func TestBasicRand(t *testing.T) {
	input := "1d4"
	output := Parse(input, harness)
	assert.Equal(t, "1 d 4 = 4", output, "rolls die but uses fake random numbers")
}

func TestVariableAssignment(t *testing.T) {
	input := "juice"
	output := Parse(input, harness)
	assert.Equal(t, "juice = 4", output, "variable 'juice' is not assigned")
}

func TestVariableMath(t *testing.T) {
	input := "juice + 3"
	output := Parse(input, harness)
	assert.Equal(t, "juice + 3 = 7", output, "variable 'juice' is not working with math")
}

func TestVariableDice(t *testing.T) {
	input := "3 d juice"
	output := Parse(input, harness)
	assert.Equal(t, "3 d juice = 12", output, "variable 'juice' is not working as die value")
}
