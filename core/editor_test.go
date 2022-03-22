package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func lines(e *Editor) [][]rune {
	lines := [][]rune{}
	for i := 0; i < e.Buffer.GetLength(); i++ {
		lines = append(lines, e.Buffer.GetLine(i))
	}
	return lines
}

func TestSortCursors(t *testing.T) {
	cursors := []*Cursor{
		{Range: Range{Location{2, 1}, Location{2, 2}}},
		{Range: Range{Location{1, 1}, Location{1, 2}}},
		{Range: Range{Location{1, 2}, Location{1, 3}}},
	}

	sorted := SortCursors(cursors)

	expected := []*Cursor{
		{Range: Range{Location{1, 1}, Location{1, 2}}},
		{Range: Range{Location{1, 2}, Location{1, 3}}},
		{Range: Range{Location{2, 1}, Location{2, 2}}},
	}

	assert.Equal(t, expected, sorted)
	assert.NotEqual(t, expected, cursors)
}

func TestUndoLimit(t *testing.T) {
	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune([]string{"0000", "1111", "2222", "3333"}))
	e := Editor{Buffer: b}
	e.Undo()
	e.Undo()
	e.Undo()
	e.Undo()
}

func TestUndo(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b}

	SetCursors(1,2,1,3)(e)
	e.MarkUndo()

	AsEdit(Split)(e)

	expected := []string{"0000", "11", "11", "2222", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())

	e.MarkUndo()
	AsEdit(Insert([]rune("!")))(e)
	expected = []string{"0000", "11!11", "2222", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestRedo(t *testing.T) {
	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune([]string{"0000", "1111", "2222", "3333"}))
	e := &Editor{Buffer: b}
	e.Redo()
	e.Redo()
	e.Redo()
	e.Redo()

	AsEdit(Split)(e)
}

func TestUndoAndRedo(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b}

	SetCursors(1,2,1,3)(e)
	e.MarkUndo()

	AsEdit(Split)(e)

	expected := []string{"0000", "11", "11", "2222", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())

	e.MarkUndo()
	AsEdit(Insert([]rune("!")))(e)
	expected = []string{"0000", "11!11", "2222", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())

	e.Redo()
	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())

	e.Redo()
	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())
}
