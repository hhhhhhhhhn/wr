package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestMoveColumns(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{1,1},Location{1,3}}),
	)
	e.Do(
		PushCursor(&Range{Location{2,3},Location{2,4}}),
	)
	e.CursorDo(MoveColumns(2))
	e.CursorDo(Split())
	e.CursorDo(MoveColumns(-100))
	e.CursorDo(Insert("!"))
	e.CursorDo(MoveColumns(-1))
	e.CursorDo(Delete())

	expected := []string{"0000", "111", "1", "2222", "", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestRemoveCursor(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{1,1},Location{1,2}}),
	)
	e.Do(
		PushCursor(&Range{Location{2,3},Location{2,4}}),
	)
	e.Do(RemoveCursor(e.Cursors[0]))
	e.CursorDo(Split())

	expected := []string{"0000", "1111", "222", "2", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestMoveRows(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{1,2},Location{1,3}}),
	)
	e.Do(
		PushCursor(&Range{Location{2,3},Location{2,4}}),
	)
	e.CursorDo(MoveRows(2))
	e.CursorDo(Split())

	expected := []string{"0000", "1111", "2222", "33", "33"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}
