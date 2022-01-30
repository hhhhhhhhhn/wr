package main

import "strings"

type Edit interface {
	Do(editor *Editor)
	Undo(editor *Editor)
	Name() string
}

// Structs in lower and Constructurs in upper
type undoMarker struct {}

func UndoMarker() undoMarker {
	return undoMarker{}
}

func (u undoMarker) Do(*Editor) {}
func (u undoMarker) Undo(*Editor) {}
func (u undoMarker) Name() string { return "Undo Marker" }

type pushCursor struct {
	_range *Range
}

func PushCursor(_range *Range) pushCursor {
	return pushCursor{_range}
}

func (p pushCursor) Do(e *Editor) {
	e.Cursors = append(e.Cursors, p._range)
}

func (p pushCursor) Undo(e *Editor) {
	e.Cursors = e.Cursors[:len(e.Cursors) - 1]
}

func (p pushCursor) Name() string {
	return "Push Cursor"
}

// Used internally, for converting cursorEdits into Edits
type cursorEditWrapper struct {
	cursorEdit CursorEdit
}

func wrapCursorEdit(cursorEdit CursorEdit) cursorEditWrapper {
	return cursorEditWrapper{cursorEdit}
}

func (c cursorEditWrapper) Do(e *Editor) {
	sortedCursors := SortCursors(e.Cursors)
	// CursorEdits are done in reverse cursor order, as, for example,
	// inserting a line messes with the location of the cursors
	// following the insert.
	for i := len(sortedCursors) - 1; i >= 0; i-- {
		c.cursorEdit.Do(e, sortedCursors[i])
	}
}

func (c cursorEditWrapper) Undo(e *Editor) {
	sortedCursors := SortCursors(e.Cursors)
	for _, cursor := range sortedCursors {
		c.cursorEdit.Undo(e, cursor)
	}
}

func (c cursorEditWrapper) Name() string {
	return c.cursorEdit.Name()
}

type CursorEdit interface {
	Do(editor *Editor, cursor *Range)
	Undo(editor *Editor, cursor *Range)
	Name() string
}

type split struct {
	originalCursors map[*Range]Range
}

func Split() split {
	return split{make(map[*Range]Range)}
}

func (s split) Do(editor *Editor, cursor *Range) {
	s.originalCursors[cursor] = *cursor

	line := editor.Buffer.GetLine(cursor.Start.Row)
	lineIndex := LocationToLineIndex(editor, cursor.Start)

	line1 := line[:lineIndex]
	line2 := line[lineIndex:]

	editor.Buffer.ChangeLine(cursor.Start.Row, line1)
	editor.Buffer.AddLine(cursor.Start.Row + 1, line2)

	*cursor = Range{Location{cursor.Start.Row + 1, 0}, Location{cursor.Start.Row+1, 1}}
}

func (s split) Undo(editor *Editor, cursor *Range) {
	line1 := editor.Buffer.GetLine(cursor.Start.Row - 1)
	line2 := editor.Buffer.GetLine(cursor.Start.Row)

	editor.Buffer.ChangeLine(cursor.Start.Row - 1, line1 + line2)
	editor.Buffer.RemoveLine(cursor.Start.Row)

	*cursor = s.originalCursors[cursor]
}

func (s split) Name() string {
	return "Split"
}

// Used by insert
type insertInLine struct {
	originalCursors map[*Range]Range
	originalLines   map[*Range]string
	insertion       string
}

func InsertInLine(insertion string) insertInLine {
	return insertInLine{
		make(map[*Range]Range), 
		make(map[*Range]string),
		insertion,
	}
}

func (s insertInLine) Do(editor *Editor, cursor *Range) {
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

func (s insertInLine) Undo(editor *Editor, cursor *Range) {
	editor.Buffer.ChangeLine(cursor.Start.Row, s.originalLines[cursor])
	*cursor = s.originalCursors[cursor]
}

func (s insertInLine) Name() string {
	return "Insert In Line"
}

type insert struct {
	edits []CursorEdit
}

func Insert(insertion string) insert {
	edits := []CursorEdit{}
	for i, line := range strings.Split(insertion, "\n") {
		if i > 0 {
			edits = append(edits, Split())
		}
		edits = append(edits, InsertInLine(line))
	}
	return insert{edits}
}

func (s insert) Do(editor *Editor, cursor *Range) {
	for _, edit := range s.edits {
		edit.Do(editor, cursor)
	}
}

func (s insert) Undo(editor *Editor, cursor *Range) {
	for i := len(s.edits) - 1; i >= 0; i-- {
		s.edits[i].Undo(editor, cursor)
	}
}

func (s insert) Name() string {
	return "Insert"
}

type delete struct {
	originalCursors map[*Range]Range
	originalLines map[*Range][]string
}

func Delete() delete {
	return delete{make(map[*Range]Range), make(map[*Range][]string)}
}

func (d delete) Do(editor *Editor, cursor *Range) {
	d.originalCursors[cursor] = *cursor

	cursorIncludeNewline(editor, cursor)

	originalLines := []string{editor.Buffer.GetLine(cursor.Start.Row)}

	for lineNumber := cursor.Start.Row + 1; lineNumber <= cursor.End.Row; lineNumber++ {
		originalLines = append(originalLines, editor.Buffer.GetLine(lineNumber))
	}

	d.originalLines[cursor] = originalLines

	cursorStartIndex := LocationToLineIndex(editor, cursor.Start)
	cursorEndIndex := LocationToLineIndex(editor, cursor.End)

	newLine := 
		originalLines[0][:cursorStartIndex] +
		originalLines[len(originalLines)-1][cursorEndIndex:]

	for lineNumber := cursor.End.Row; lineNumber >= cursor.Start.Row + 1; lineNumber-- {
		// NOTE: This is very inneficient, as all lines are relocated after
		// every delete
		editor.Buffer.RemoveLine(lineNumber)
	}
	editor.Buffer.ChangeLine(cursor.Start.Row, newLine)

	cursor.End.Row = cursor.Start.Row
	cursor.End.Column = cursor.Start.Column+1
}

func (d delete) Undo(editor *Editor, cursor *Range) {
	editor.Buffer.ChangeLine(cursor.Start.Row, d.originalLines[cursor][0])

	lineNumber := cursor.Start.Row + 1
	i := 1

	for i < len(d.originalLines[cursor]) {
		editor.Buffer.AddLine(lineNumber, d.originalLines[cursor][i])
		lineNumber++
		i++
	}

	*cursor = d.originalCursors[cursor]
}

func (d delete) Name() string {
	return "Delete"
}

// By default, the cursor can be one more character to the right than there is
// in the line, representing the newline and allowing for insertions after the
// last character. This means the cursor end is OOB.
// This function simply moves those cursors ends into the start of the next line,
func cursorIncludeNewline(editor *Editor, cursor *Range) {
	if !isInBounds(editor, cursor.End) {
		cursor.End.Row++
		cursor.End.Column = 0
	}
}

func isInBounds(editor *Editor, location Location) bool {
	if location.Column == 0 {
		return true
	}

	location.Column--
	line := editor.Buffer.GetLine(location.Row)
	if StringColumnSpan(editor, line) == location.Column {
		return false
	}
	return true
}
