package main

import "sort"

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
		if e.HistoryIndex == 0 {
			break
		}
		e.HistoryIndex--
		e.History[e.HistoryIndex].Undo(e)
		if e.History[e.HistoryIndex].Name() == "Undo Marker" {
			break
		}
	}
}

func (e *Editor) Redo() {
	for {
		if e.HistoryIndex == len(e.History) - 1 {
			break
		}
		e.History[e.HistoryIndex].Do(e)
		if e.History[e.HistoryIndex].Name() == "Undo Marker" {
			break
		}
		e.HistoryIndex++
	}
}

func (e *Editor) CursorDo(cursorEdit CursorEdit) {
	e.Do(wrapCursorEdit(cursorEdit))
}

func LocationToLineIndex(editor *Editor, location Location) int {
	index :=  0
	column := 0
	line := editor.Buffer.GetLine(location.Row)
	for column < location.Column {
		switch(line[index]) {
		case '\t':
			column += editor.Config.Tabsize
		default:
			column++
		}
		index++
	}
	return index
}

func LineIndexToColumn(editor *Editor, index int, line string) int {
	i := 0
	column := 0
	for i < index {
		switch(line[index]) {
		case '\t':
			column += editor.Config.Tabsize
		default:
			column++
		}
		i++
	}
	return column
}

func StringColumnSpan(editor *Editor, str string) int {
	index := 0
	column := 0
	for index < len(str) {
		switch(str[index]) {
		case '\t':
			column += editor.Config.Tabsize
		default:
			column++
		}
		index++
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

