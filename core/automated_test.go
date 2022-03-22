// +build !noAuto

package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type scenario struct {
	lines   []string
	cursors []Cursor
}

var scenarios = []scenario {
	{
		[]string{"aaaa", "bbbb", "cccc", "dddd"},
		[]Cursor{
			{Range: Range{Location{1, 1}, Location{2, 2}}},
		},
	},
	{
		[]string{"aaaa", "bbbb", "cccc", "dddd"},
		[]Cursor{
			{Range: Range{Location{1, 1}, Location{1, 4}}},
			{Range: Range{Location{2, 1}, Location{2, 4}}},
			{Range: Range{Location{3, 1}, Location{3, 4}}},
		},
	},
}

var cursorEdits = []func()CursorEdit{
	func() CursorEdit {return Split},
	func() CursorEdit {return Insert([]rune("!"))},
	func() CursorEdit {return Insert([]rune("\n"))},
	func() CursorEdit {return Delete},
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
	originalCursors := append([]Cursor{}, s.cursors...)

	buffer := NewBuffer()
	buffer.Current = buffer.Current.Insert(0, originalLines)
	editor := &Editor{Buffer: buffer}

	for _, cursor := range s.cursors {
		cursorCopy := cursor
		PushCursor(&cursorCopy)(editor)
	}

	editor.MarkUndo()

	AsEdit(ce())(editor)
	updatedLines := lines(editor)
	updatedCursors := cursors(editor)

	editor.Undo()
	assert.Equal(t, originalLines, editor.Buffer.Current.Value())
	assert.Equal(t, originalCursors, cursors(editor))

	editor.Redo()
	assert.Equal(t, updatedLines, editor.Buffer.Current.Value())
	assert.Equal(t, updatedCursors, cursors(editor))

	editor.Undo()
	assert.Equal(t, originalLines, editor.Buffer.Current.Value())
	assert.Equal(t, originalCursors, cursors(editor))

	editor.Redo()
	assert.Equal(t, updatedLines, editor.Buffer.Current.Value())
	assert.Equal(t, updatedCursors, cursors(editor))


	originalLines = CopyLines(updatedLines)
	originalCursors = append([]Cursor{}, updatedCursors...)

	editor.MarkUndo()

	AsEdit(ce())(editor)
	updatedLines = lines(editor)
	updatedCursors = cursors(editor)

	editor.Undo()
	assert.Equal(t, originalLines, editor.Buffer.Current.Value())
	assert.Equal(t, originalCursors, cursors(editor))

	editor.Redo()
	assert.Equal(t, updatedLines, editor.Buffer.Current.Value())
	assert.Equal(t, updatedCursors, cursors(editor))

	editor.Undo()
	assert.Equal(t, originalLines, editor.Buffer.Current.Value())
	assert.Equal(t, originalCursors, cursors(editor))

	editor.Redo()
	assert.Equal(t, updatedLines, editor.Buffer.Current.Value())
	assert.Equal(t, updatedCursors, cursors(editor))

	editor.Undo()
	editor.Undo()
}

func cursors(editor *Editor) []Cursor {
	dst := []Cursor{}
	for _, cursor := range editor.Cursors {
		cursorCopy := *cursor
		dst = append(dst, cursorCopy)
	}
	return dst
}
