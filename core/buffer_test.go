package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestAddLine(t *testing.T) {
	b := Buffer{}
	b.AddLine(0, []rune("Line 1"))
	b.AddLine(1, []rune("Line 3"))
	b.AddLine(1, []rune("Line 2"))
	
	expected := [][]rune{
		[]rune("Line 1"),
		[]rune("Line 2"),
		[]rune("Line 3"),
	}
	assert.Equal(t, expected, b.Lines)
}

func TestRemoveLine(t *testing.T) {
	b := Buffer{}
	b.AddLine(0, []rune("Line 1"))
	b.AddLine(1, []rune("Line 3"))
	b.AddLine(1, []rune("Line 2"))

	b.RemoveLine(1)
	expected := []string{"Line 1", "Line 3"}
	assert.Equal(t, b.Lines, ToRune(expected))

	b.RemoveLine(0)
	expected = []string{"Line 3"}
	assert.Equal(t, b.Lines, ToRune(expected))

	b.RemoveLine(0)
	expected = []string{}
	assert.Equal(t, b.Lines, ToRune(expected))
}

func TestChangeLine(t *testing.T) {
	b := Buffer{}
	b.AddLine(0, []rune("Line 1"))
	b.AddLine(1, []rune("Line 3"))
	b.AddLine(1, []rune("Line 2"))

	b.ChangeLine(1, []rune("New Line 2"))
	expected := []string{"Line 1", "New Line 2", "Line 3"}
	assert.Equal(t, b.Lines, ToRune(expected))
}
