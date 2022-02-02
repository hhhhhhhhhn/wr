package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSortCursors(t *testing.T) {
	cursors := []*Range{
		{Location{2, 1}, Location{2, 2}},
		{Location{1, 1}, Location{1, 2}},
		{Location{1, 2}, Location{1, 3}},
	}

	sorted := SortCursors(cursors)

	expected := []*Range{
		{Location{1, 1}, Location{1, 2}},
		{Location{1, 2}, Location{1, 3}},
		{Location{2, 1}, Location{2, 2}},
	}

	assert.Equal(t, expected, sorted)
	assert.NotEqual(t, expected, cursors)
}

func TestUndoLimit(t *testing.T) {
	b := Buffer{Lines: []string{"0000", "1111", "2222", "3333"}}
	e := Editor{Buffer: &b}
	e.Undo()
	e.Undo()
	e.Undo()
	e.Undo()
}

func TestUndo(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b}

	e.Do(
		PushCursor(&Range{Location{1,2},Location{1,3}}),
	)
	e.Do(UndoMarker())
	e.CursorDo(Split())

	expected := []string{"0000", "11", "11", "2222", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)

	e.Do(UndoMarker())
	e.CursorDo(Insert("!"))
	expected = []string{"0000", "11!11", "2222", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}
