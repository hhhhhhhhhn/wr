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

var cursorEdits = []func()CursorEdit{
	func() CursorEdit {return Split()},
	func() CursorEdit {return Insert([]rune("!"))},
	func() CursorEdit {return Insert([]rune("\n"))},
	func() CursorEdit {return Delete()},
	func() CursorEdit {return MoveColumns(1)},
	func() CursorEdit {return MoveColumns(10)},
	func() CursorEdit {return MoveColumns(-1)},
	func() CursorEdit {return MoveColumns(-10)},
	func() CursorEdit {return MoveRows(1)},
	func() CursorEdit {return MoveRows(10)},
	func() CursorEdit {return MoveRows(-1)},
	func() CursorEdit {return MoveRows(-10)},
	func() CursorEdit {return MoveChars(1)},
	func() CursorEdit {return MoveChars(-1)},
	func() CursorEdit {return MoveChars(100)},
	func() CursorEdit {return MoveChars(-100)},
	func() CursorEdit {return BoundToLine()},
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

func testScenario(t *testing.T, s scenario, ce func() CursorEdit) {
	originalLines := ToRune(s.lines)
	originalCursors := append([]Range{}, s.cursors...)

	buffer := &Buffer{Lines: CopyLines(originalLines)}
	editor := Editor{Buffer: buffer}

	for _, cursor := range s.cursors {
		cursorCopy := cursor
		editor.Do(PushCursor(&cursorCopy))
	}

	editor.Do(UndoMarker())

	editor.CursorDo(ce())
	updatedLines := lines(&editor)
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


	originalLines = CopyLines(updatedLines)
	originalCursors = append([]Range{}, updatedCursors...)

	editor.Do(UndoMarker())

	editor.CursorDo(ce())
	updatedLines = lines(&editor)
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
