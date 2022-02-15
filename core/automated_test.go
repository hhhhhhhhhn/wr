// +build !noAuto

package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type scenario struct {
	lines   []string
	cursors []Range
}

var scenarios = []scenario {
	{
		[]string{"aaaa", "bbbb", "cccc", "dddd"},
		[]Range{
			{Location{1, 1}, Location{2, 2}},
		},
	},
	{
		[]string{"aaaa", "bbbb", "cccc", "dddd"},
		[]Range{
			{Location{1, 1}, Location{1, 4}},
			{Location{2, 1}, Location{2, 4}},
			{Location{3, 1}, Location{3, 4}},
		},
	},
}

var cursorEdits = []CursorEdit{
	Split(),
	Insert("!"),
	Insert("\n"),
	Delete(),
	MoveColumns(1),
	MoveColumns(10),
	MoveColumns(-1),
	MoveColumns(-10),
	MoveRows(1),
	MoveRows(10),
	MoveRows(-1),
	MoveRows(-10),
	MoveChars(1),
	MoveChars(-1),
	MoveChars(100),
	MoveChars(-100),
	BoundToLine(),
}

func TestScenarios(t *testing.T) {
	for i, cursorEdit := range cursorEdits {
		t.Log("CursorEdit:", i)
		for j, scenario := range scenarios {
			t.Log("Scenario:", j)
			testScenario(t, scenario, cursorEdit)
		}
	}
}

func testScenario(t *testing.T, s scenario, ce CursorEdit) {
	originalLines := append([]string{}, s.lines...)
	originalCursors := append([]Range{}, s.cursors...)

	buffer := &Buffer{Lines: append([]string{}, originalLines...)}
	editor := Editor{Buffer: buffer}

	for _, cursor := range s.cursors {
		editor.Do(PushCursor(&cursor))
	}

	editor.Do(UndoMarker())

	editor.CursorDo(ce)
	updatedLines := append([]string{}, editor.Buffer.Lines...)
	updatedCursors := cursors(&editor)

	editor.Undo()
	assert.Equal(t, originalLines, editor.Buffer.Lines)
	assert.Equal(t, originalCursors, cursors(&editor))

	editor.Redo()
	assert.Equal(t, updatedLines, editor.Buffer.Lines)
	assert.Equal(t, updatedCursors, cursors(&editor))

	editor.Undo()
	assert.Equal(t, originalLines, editor.Buffer.Lines)
	assert.Equal(t, originalCursors, cursors(&editor))

	editor.Redo()
	assert.Equal(t, updatedLines, editor.Buffer.Lines)
	assert.Equal(t, updatedCursors, cursors(&editor))


	originalLines = append([]string{}, updatedLines...)
	originalCursors = append([]Range{}, updatedCursors...)

	editor.Do(UndoMarker())

	editor.CursorDo(ce)
	updatedLines = append([]string{}, editor.Buffer.Lines...)
	updatedCursors = cursors(&editor)

	editor.Undo()
	assert.Equal(t, originalLines, editor.Buffer.Lines)
	assert.Equal(t, originalCursors, cursors(&editor))

	editor.Redo()
	assert.Equal(t, updatedLines, editor.Buffer.Lines)
	assert.Equal(t, updatedCursors, cursors(&editor))

	editor.Undo()
	assert.Equal(t, originalLines, editor.Buffer.Lines)
	assert.Equal(t, originalCursors, cursors(&editor))

	editor.Redo()
	assert.Equal(t, updatedLines, editor.Buffer.Lines)
	assert.Equal(t, updatedCursors, cursors(&editor))

	editor.Undo()
	editor.Undo()
}

func cursors(editor *Editor) []Range {
	dst := []Range{}
	for _, cursor := range editor.Cursors {
		dst = append(dst, *cursor)
	}
	return dst
}
