package core

import (
	"sort"
	"math/rand"
	"io"
	"os"

	rw "github.com/mattn/go-runewidth"
)

type Cursor struct {
	Range
	Registers []string
}

type Editor struct {
	Buffer          *Buffer
	History         []Version
	HistoryIndex    int
	Cursors         []*Cursor
	CursorsVersions map[Version][]Cursor
	Config          EditorConfig
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

func (e *Editor) Undo() {
	if e.HistoryIndex >= len(e.History) { // Current version is not saved
		e.MarkUndo()
		e.HistoryIndex--
	}
	if e.HistoryIndex > 0 {
		e.HistoryIndex--
	}
	if e.HistoryIndex < len(e.History) {
		e.restoreVersion(e.History[e.HistoryIndex])
	}
}

func (e *Editor) Redo() {
	if e.HistoryIndex < len(e.History) - 1 {
		e.HistoryIndex++
	}
	if e.HistoryIndex < len(e.History) {
		e.restoreVersion(e.History[e.HistoryIndex])
	}
}

func (e *Editor) restoreVersion(version Version) {
	e.Buffer.Restore(version)
	e.Cursors = restoreCursors(e.CursorsVersions[version])
}

// Marks the start of an action to be undone
func (e *Editor) MarkUndo() {
	if e.CursorsVersions == nil {
		e.CursorsVersions = make(map[int][]Cursor)
	}
	newVersion := rand.Int()
	e.Buffer.Backup(newVersion)
	e.CursorsVersions[newVersion] = backupCursors(e.Cursors)
	e.HistoryIndex++
	e.History = append(e.History[:e.HistoryIndex-1], newVersion)
}

func backupCursors(cursors []*Cursor) []Cursor {
	backup := make([]Cursor, len(cursors))
	for i, cursor := range cursors {
		backup[i] = *cursor
	}
	return backup
}

func restoreCursors(cursors []Cursor) []*Cursor {
	restored := make([]*Cursor, len(cursors))
	for i, cursor := range cursors {
		cursorCopy := cursor
		restored[i] = &cursorCopy
	}
	return restored
}

func RuneWidth(editor *Editor, chr rune) int {
	switch(chr) {
	case '\t':
		return editor.Config.Tabsize
	default:
		return rw.RuneWidth(chr)
	}
}

func LocationToIndex(editor *Editor, location Location) int {
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

func ColumnToIndex(editor *Editor, line []rune, column int) int {
	currentCol := 0
	for i, chr := range line {
		currentCol += RuneWidth(editor, chr)
		if currentCol > column {
			return i
		}
	}
	return len(line)
}

func ColumnSpan(editor *Editor, line []rune) (column int) {
	for _, chr := range line {
		column += RuneWidth(editor, chr)
	}
	return column
}

func SortCursors(cursors []*Cursor) (sortedCursors []*Cursor) {
	sortedCursors = make([]*Cursor, len(cursors))
	copy(sortedCursors, cursors)
	sort.Slice(sortedCursors, func(i, j int) bool {
		if sortedCursors[i].Start.Row == sortedCursors[j].Start.Row {
			return sortedCursors[i].Start.Column < sortedCursors[j].Start.Column
		}
		return sortedCursors[i].Start.Row < sortedCursors[j].Start.Row
	})
	return sortedCursors
}

func ToRune(lines []string) [][]rune {
	runes := [][]rune{}
	for _, line := range lines {
		runes = append(runes, []rune(line))
	}
	return runes
}

func CopyLines(lines [][]rune) [][]rune {
	copied := make([][]rune, len(lines))
	for i, line := range lines {
		copied[i] = make([]rune, len(line))
		copy(copied[i], line)
	}
	return copied
}

func SaveToFile(editor *Editor, filename string) error {
	reader := NewEditorReader(editor, 0, 0)
	data, err:= io.ReadAll(reader)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
