package core

import (
	"sort"
	rw "github.com/mattn/go-runewidth"
)

type Editor struct {
	Buffer       *Buffer
	History      []Edit
	HistoryIndex int
	Cursors      []*Range
	Config       EditorConfig
}

type EditorConfig struct {
	Tabsize int
}

type Range struct {
	Start Location
	End   Location
}

// NOTE: For simplicity, locations are zero-indexed.
type Location struct {
	Row    int
	Column int
}

func (e *Editor) Do(edit Edit) {
	e.History = append(e.History[:e.HistoryIndex], edit)
	e.HistoryIndex++
	edit.Do(e)
}

func (e *Editor) Undo() {
	for {
		e.HistoryIndex--
		if e.HistoryIndex < 0 {
			e.Redo()
			break
		}
		e.History[e.HistoryIndex].Undo(e)
		if e.History[e.HistoryIndex].Name() == "Undo Marker" {
			break
		}
	}
}

func (e *Editor) Redo() {
	for {
		e.HistoryIndex++
		if e.HistoryIndex > len(e.History) - 1 {
			e.HistoryIndex = len(e.History)
			break
		}
		e.History[e.HistoryIndex].Do(e)
		if e.History[e.HistoryIndex].Name() == "Undo Marker" {
			break
		}
	}
}

func (e *Editor) CursorDo(cursorEdit CursorEdit) {
	e.Do(wrapCursorEdit(cursorEdit))
}

func RuneWidth(editor *Editor, chr rune) int {
	switch(chr) {
	case '\t':
		return editor.Config.Tabsize
	default:
		return rw.RuneWidth(chr)
	}
}

func LocationToByteIndex(editor *Editor, location Location) int {
	column := 0
	line := editor.Buffer.GetLine(location.Row)
	for i, chr := range line {
		if column >= location.Column {
			return i
		}
		column += RuneWidth(editor, chr)
	}
	return len(line)
}

func ByteIndexToColumn(editor *Editor, index int, line string) int {
	column := 0
	for i, chr := range line {
		if i >= index {
			break
		}
		column += RuneWidth(editor, chr)
	}
	return column
}

func StringColumnSpan(editor *Editor, str string) int {
	column := 0
	for _, chr := range str {
		column += RuneWidth(editor, chr)
	}
	return column
}

func SortCursors(cursors []*Range) (sortedCursors []*Range) {
	sortedCursors = make([]*Range, len(cursors))
	copy(sortedCursors, cursors)
	sort.Slice(sortedCursors, func(i, j int) bool {
		if sortedCursors[i].Start.Row == sortedCursors[j].Start.Row {
			return sortedCursors[i].Start.Column < sortedCursors[j].Start.Column
		}
		return sortedCursors[i].Start.Row < sortedCursors[j].Start.Row
	})
	return sortedCursors
}

