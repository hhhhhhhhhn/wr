package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestAddLine(t *testing.T) {
	b := Buffer{}
	b.AddLine(0, "Line 1")
	b.AddLine(1, "Line 3")
	b.AddLine(1, "Line 2")
	
	expected := []string{"Line 1", "Line 2", "Line 3"}
	assert.Equal(t, b.Lines, expected)
}

func TestRemoveLine(t *testing.T) {
	b := Buffer{Lines: []string{"Line 1", "Line 2", "Line 3"}}

	b.RemoveLine(1)
	expected := []string{"Line 1", "Line 3"}
	assert.Equal(t, b.Lines, expected)

	b.RemoveLine(0)
	expected = []string{"Line 3"}
	assert.Equal(t, b.Lines, expected)

	b.RemoveLine(0)
	expected = []string{}
	assert.Equal(t, b.Lines, expected)
}

func TestChangeLine(t *testing.T) {
	b := Buffer{Lines: []string{"Line 1", "Line 2", "Line 3"}}

	b.ChangeLine(1, "New Line 2")
	expected := []string{"Line 1", "New Line 2", "Line 3"}
	assert.Equal(t, b.Lines, expected)
}
