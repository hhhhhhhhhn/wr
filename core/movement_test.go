package core

import (
	"testing"
)

func TestGoToUp(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b}
	e.MarkUndo()

	SetCursors(0, 0, 0, 1)(e)

	GoTo(Rows(-1))(e)
	GoTo(Rows(-1))(e)
}
