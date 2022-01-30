package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSplit(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b}

	e.AddEdit(
		NewUndoMarkerEdit(),
	)
	e.AddEdit(
		NewPushCursorEdit(&Range{Location{1,2},Location{1,3}}),
	)
	e.AddEdit(
		NewPushCursorEdit(&Range{Location{2,3},Location{2,4}}),
	)
	e.AddCursorEdit(NewSplitCursorEdit())

	expected := []string{"0000", "11", "11", "222", "2", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestInsertInLine(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.AddEdit(
		NewUndoMarkerEdit(),
	)
	e.AddEdit(
		NewPushCursorEdit(&Range{Location{1,2},Location{1,3}}),
	)
	e.AddEdit(
		NewPushCursorEdit(&Range{Location{2,3},Location{2,4}}),
	)
	e.AddCursorEdit(newInsertInLineCursorEdit("!\t"))
	e.AddCursorEdit(newInsertInLineCursorEdit("!"))

	expected := []string{"0000", "11!\t!11", "222!\t!2", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestInsert(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.AddEdit(
		NewUndoMarkerEdit(),
	)
	e.AddEdit(
		NewPushCursorEdit(&Range{Location{1,2},Location{1,3}}),
	)
	e.AddEdit(
		NewPushCursorEdit(&Range{Location{2,3},Location{2,4}}),
	)
	e.AddCursorEdit(NewInsertCursorEdit("!\n!"))

	expected := []string{"0000", "11!", "!11", "222!", "!2", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}
