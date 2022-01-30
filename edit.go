package main

import "strings"

type Edit interface {
	Do(editor *Editor)
	Undo(editor *Editor)
	Name() string
}

type UndoMarkerEdit struct {}

func NewUndoMarkerEdit() UndoMarkerEdit {
	return UndoMarkerEdit{}
}

func (u UndoMarkerEdit) Do(*Editor) {}
func (u UndoMarkerEdit) Undo(*Editor) {}
func (u UndoMarkerEdit) Name() string { return "Undo Marker" }

type PushCursorEdit struct {
	_range *Range
}

func NewPushCursorEdit(_range *Range) PushCursorEdit {
	return PushCursorEdit{_range}
}

func (p PushCursorEdit) Do(e *Editor) {
	e.Cursors = append(e.Cursors, p._range)
}

func (p PushCursorEdit) Undo(e *Editor) {
	e.Cursors = e.Cursors[:len(e.Cursors) - 1]
}

func (p PushCursorEdit) Name() string {
	return "Push Cursor"
}

// Used internally, for converting cursorEdits into Edits
type cursorEditEdit struct {
	cursorEdit CursorEdit
}

func newCursorEditEdit(cursorEdit CursorEdit) cursorEditEdit {
	return cursorEditEdit{cursorEdit}
}

func (c cursorEditEdit) Do(e *Editor) {
	sortedCursors := SortCursors(e.Cursors)
	// CursorEdits are done in reverse cursor order, as, for example,
	// inserting a line messes with the location of the cursors
	// following the insert.
	for i := len(sortedCursors) - 1; i >= 0; i-- {
		c.cursorEdit.Do(e, sortedCursors[i])
	}
}

func (c cursorEditEdit) Undo(e *Editor) {
	sortedCursors := SortCursors(e.Cursors)
	for _, cursor := range sortedCursors {
		c.cursorEdit.Undo(e, cursor)
	}
}

func (c cursorEditEdit) Name() string {
	return c.cursorEdit.Name()
}

type CursorEdit interface {
	Do(editor *Editor, cursor *Range)
	Undo(editor *Editor, cursor *Range)
	Name() string
}

type SplitCursorEdit struct {
	originalCursors map[*Range]Range
}

func NewSplitCursorEdit() SplitCursorEdit {
	return SplitCursorEdit{make(map[*Range]Range)}
}

func (s SplitCursorEdit) Do(editor *Editor, cursor *Range) {
	s.originalCursors[cursor] = *cursor

	line := editor.Buffer.GetLine(cursor.Start.Row)
	lineIndex := LocationToLineIndex(editor, cursor.Start)

	line1 := line[:lineIndex]
	line2 := line[lineIndex:]

	editor.Buffer.ChangeLine(cursor.Start.Row, line1)
	editor.Buffer.AddLine(cursor.Start.Row + 1, line2)

	*cursor = Range{Location{cursor.Start.Row + 1, 0}, Location{cursor.Start.Row+1, 1}}
}

func (s SplitCursorEdit) Undo(editor *Editor, cursor *Range) {
	line1 := editor.Buffer.GetLine(cursor.Start.Row - 1)
	line2 := editor.Buffer.GetLine(cursor.Start.Row)

	editor.Buffer.ChangeLine(cursor.Start.Row - 1, line1 + line2)
	editor.Buffer.RemoveLine(cursor.Start.Row)

	*cursor = s.originalCursors[cursor]
}

func (s SplitCursorEdit) Name() string {
	return "Split"
}

// Used by insert
type insertInLineCursorEdit struct {
	originalCursors map[*Range]Range
	originalLines   map[*Range]string
	insertion       string
}

func newInsertInLineCursorEdit(insertion string) insertInLineCursorEdit {
	return insertInLineCursorEdit{
		make(map[*Range]Range), 
		make(map[*Range]string),
		insertion,
	}
}

func (s insertInLineCursorEdit) Do(editor *Editor, cursor *Range) {
	s.originalCursors[cursor] = *cursor
	
	line := editor.Buffer.GetLine(cursor.Start.Row)
	s.originalLines[cursor] = line

	lineIndex := LocationToLineIndex(editor, cursor.Start)

	newLine := line[:lineIndex] + s.insertion + line[lineIndex:]

	editor.Buffer.ChangeLine(cursor.Start.Row, newLine)

	newCursorStart := cursor.Start
	newCursorStart.Column += StringColumnSpan(editor, s.insertion)
	newCursorEnd := newCursorStart
	newCursorEnd.Column += 1

	cursor.Start = newCursorStart
	cursor.End = newCursorEnd
}

func (s insertInLineCursorEdit) Undo(editor *Editor, cursor *Range) {
	editor.Buffer.ChangeLine(cursor.Start.Row, s.originalLines[cursor])
	*cursor = s.originalCursors[cursor]
}

func (s insertInLineCursorEdit) Name() string {
	return "Insert In Line"
}

type InsertCursorEdit struct {
	edits []CursorEdit
}

func NewInsertCursorEdit(insertion string) InsertCursorEdit {
	edits := []CursorEdit{}
	for i, line := range strings.Split(insertion, "\n") {
		if i > 0 {
			edits = append(edits, NewSplitCursorEdit())
		}
		edits = append(edits, newInsertInLineCursorEdit(line))
	}
	return InsertCursorEdit{edits}
}

func (s InsertCursorEdit) Do(editor *Editor, cursor *Range) {
	for _, Edit := range s.edits {
		Edit.Do(editor, cursor)
	}
}

func (s InsertCursorEdit) Undo(editor *Editor, cursor *Range) {
	for i := len(s.edits) - 1; i >= 0; i-- {
		s.edits[i].Undo(editor, cursor)
	}
}

func (s InsertCursorEdit) Name() string {
	return "Insert"
}
