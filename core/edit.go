package core

type Edit interface {
	Do(editor *Editor)
	Undo(editor *Editor)
	Name() string
}
// Structs in lower and Constructurs in upper
type undoMarker struct {}

func UndoMarker() *undoMarker {
	return &undoMarker{}
}

func (u *undoMarker) Do(*Editor) {}
func (u *undoMarker) Undo(*Editor) {}
func (u *undoMarker) Name() string { return "Undo Marker" }

// Used internally, for converting cursorEdits into Edits
type cursorEditWrapper struct {
	cursorEdit    CursorEdit
	sortedCursors []*Range
	removeCursors []*removeCursor
}

func wrapCursorEdit(cursorEdit CursorEdit) *cursorEditWrapper {
	return &cursorEditWrapper{cursorEdit, nil, nil}
}

func (c *cursorEditWrapper) Do(e *Editor) {
	c.sortedCursors = SortCursors(e.Cursors)
	// CursorEdits are done in reverse cursor order, as, for example,
	// inserting a line messes with the location of the cursors
	// following the insert.
	for i := len(c.sortedCursors) - 1; i >= 0; i-- {
		c.cursorEdit.Do(e, c.sortedCursors[i], i)
	}

	lastRow := e.Buffer.GetLength() - 1
	for _, cursor := range e.Cursors {
		if cursor.Start.Row < 0 ||
			cursor.End.Row < 0 ||
			cursor.Start.Row > lastRow ||
			cursor.End.Row > lastRow {
				removeCursor := RemoveCursor(cursor)
				removeCursor.Do(e)
				c.removeCursors = append(c.removeCursors, removeCursor)
			}
	}
}

func (c *cursorEditWrapper) Undo(e *Editor) {
	for i := len(c.removeCursors) - 1; i >= 0; i-- {
		c.removeCursors[i].Undo(e)
	}

	for i, cursor := range c.sortedCursors {
		c.cursorEdit.Undo(e, cursor, i)
	}

	c.removeCursors = nil
}

func (c *cursorEditWrapper) Name() string {
	return c.cursorEdit.Name()
}

type CursorEdit interface {
	Do(editor *Editor, cursor *Range, index int)
	Undo(editor *Editor, cursor *Range, index int)
	Name() string
}

type singleSplit struct {
	originalCursors []Range
	row             int
	column          int
}

func SingleSplit(row, column int) *singleSplit {
	return &singleSplit{nil, row, column}
}

func (s *singleSplit) Do(editor *Editor) {
	s.originalCursors = copyCursors(editor)

	line := editor.Buffer.GetLine(s.row)
	lineIndex := LocationToIndex(editor, Location{s.row, s.column})

	line1 := line[:lineIndex]
	line2 := line[lineIndex:]

	editor.Buffer.ChangeLine(s.row, line1)
	editor.Buffer.AddLine(s.row + 1, line2)

	for _, cursor := range editor.Cursors {
		if cursor.Start.Row > s.row {
			cursor.Start.Row++
			cursor.End.Row++
		} else if cursor.Start.Row == s.row && cursor.Start.Column >= s.column {
			cursor.Start.Column -= s.column
			cursor.End.Column -= s.column
			cursor.Start.Row++
			cursor.End.Row++
		}
	}
}

func (s *singleSplit) Undo(editor *Editor) {
	restoreCursors(editor, s.originalCursors)

	line1 := editor.Buffer.GetLine(s.row)
	line2 := editor.Buffer.GetLine(s.row+1)

	editor.Buffer.ChangeLine(s.row, append(line1, line2...))
	editor.Buffer.RemoveLine(s.row+1)
}

func (s *singleSplit) Name() string {
	return "Single Split"
}

func copyCursors(editor *Editor) []Range {
	cursors := make([]Range, len(editor.Cursors))
	for i, cursor := range editor.Cursors {
		cursors[i] = *cursor
	}
	return cursors
}

func restoreCursors(editor *Editor, src []Range) {
	for i, cursor := range src {
		*editor.Cursors[i] = cursor
	}
}

type split struct {
	singleSplits []*singleSplit
}

func Split() *split {
	return &split{}
}

func (s *split) Do(editor *Editor, cursor *Range, index int) {
	// First one to be executed
	if index == len(editor.Cursors) - 1 {
		s.singleSplits = make([]*singleSplit, len(editor.Cursors))
	}

	singleSplit := SingleSplit(cursor.Start.Row, cursor.Start.Column)
	s.singleSplits[index] = singleSplit
	singleSplit.Do(editor)
}

func (s *split) Undo(editor *Editor, cursor *Range, index int) {
	s.singleSplits[index].Undo(editor)
}

func (s *split) Name() string {
	return "Split"
}

// Used by insertInLine
type singleInsertInLine struct {
	originalCursors []Range
	originalLine    []rune
	insertion       []rune
	row             int
	column          int
}

func SingleInsertInLine(insertion []rune, row, column int) *singleInsertInLine {
	return &singleInsertInLine{nil,nil, insertion, row, column}
}

func Join(parts... []rune) []rune {
	length := 0
	for _, part := range parts {
		length += len(part)
	}

	joined := make([]rune, length)
	i := 0

	for _, part := range parts {
		i += copy(joined[i:], part)
	}
	return joined
}

func (s *singleInsertInLine) Do(editor *Editor) {
	s.originalCursors = copyCursors(editor)

	line := editor.Buffer.GetLine(s.row)
	s.originalLine = line

	lineIndex := LocationToIndex(editor, Location{s.row, s.column})

	newLine := Join(line[:lineIndex], s.insertion, line[lineIndex:])

	editor.Buffer.ChangeLine(s.row, newLine)

	insertedColumns := ColumnSpan(editor, s.insertion)

	for _, cursor := range editor.Cursors {
		if cursor.Start.Row == s.row && cursor.Start.Column >= s.column {
			cursor.Start.Column += insertedColumns
		}
		if cursor.End.Row == s.row && cursor.End.Column >= s.column {
			cursor.End.Column += insertedColumns
		}
	}
}

func (s *singleInsertInLine) Undo(editor *Editor) {
	editor.Buffer.ChangeLine(s.row, s.originalLine)
	restoreCursors(editor, s.originalCursors)
}

func (s *singleInsertInLine) Name() string {
	return "Single Insert In Line"
}

// Used by insert
type insertInLine struct {
	singleInsertInLines []*singleInsertInLine
	insertion           []rune
}

func InsertInLine(insertion []rune) *insertInLine {
	return &insertInLine{nil, insertion}
}

func (s *insertInLine) Do(editor *Editor, cursor *Range, index int) {
	if index == len(editor.Cursors) - 1 {
		s.singleInsertInLines = make([]*singleInsertInLine, len(editor.Cursors))
	}
	singleInsertInLine := SingleInsertInLine(s.insertion, cursor.Start.Row, cursor.Start.Column)
	s.singleInsertInLines[index] = singleInsertInLine
	singleInsertInLine.Do(editor)
}

func (s *insertInLine) Undo(editor *Editor, cursor *Range, index int) {
	s.singleInsertInLines[index].Undo(editor)
}

func (s *insertInLine) Name() string {
	return "Insert In Line"
}

type insert struct {
	edits []CursorEdit
}

func Insert(insertion []rune) *insert {
	edits := []CursorEdit{}
	for i, line := range splitRune(insertion, '\n') {
		if i > 0 {
			edits = append(edits, Split())
		}
		edits = append(edits, InsertInLine(line))
	}
	return &insert{edits}
}

func splitRune(str []rune, div rune) (output [][]rune) {
	output = [][]rune{nil}
	for _, chr := range str {
		if chr == div {
			output = append(output, nil)
		} else {
			output[len(output)-1] = append(output[len(output)-1], chr)
		}
	}
	return output
}

func (s *insert) Do(editor *Editor, cursor *Range, index int) {
	for _, edit := range s.edits {
		edit.Do(editor, cursor, index)
	}
}

func (s *insert) Undo(editor *Editor, cursor *Range, index int) {
	for i := len(s.edits) - 1; i >= 0; i-- {
		s.edits[i].Undo(editor, cursor, index)
	}
}

func (s *insert) Name() string {
	return "Insert"
}

type singleDelete struct {
	area            Range
	originalLines   [][]rune
	originalCursors []Range
}

func SingleDelete(_range Range) *singleDelete {
	return &singleDelete{_range, nil, nil}
}

func (s *singleDelete) Do(editor *Editor) {
	s.originalCursors = copyCursors(editor)

	area := s.area

	cursorIncludeNewline(editor, &area)

	originalLines := [][]rune{editor.Buffer.GetLine(area.Start.Row)}

	for lineNumber := area.Start.Row + 1; lineNumber <= area.End.Row; lineNumber++ {
		originalLines = append(originalLines, editor.Buffer.GetLine(lineNumber))
	}

	s.originalLines = originalLines

	cursorStartIndex := LocationToIndex(editor, area.Start)
	cursorEndIndex := LocationToIndex(editor, area.End)

	newLine := Join(
		slice(originalLines[0], 0, cursorStartIndex),
		slice(originalLines[len(originalLines)-1], cursorEndIndex, -1),
	)

	for lineNumber := area.End.Row; lineNumber >= area.Start.Row + 1; lineNumber-- {
		// NOTE: This is very inneficient, as all lines are relocated after
		// every delete
		editor.Buffer.RemoveLine(lineNumber)
	}
	editor.Buffer.ChangeLine(area.Start.Row, newLine)

	// The amount of deleted columns on the last line
	var deletedColumns int
	if area.Start.Row == area.End.Row {
		deletedColumns = area.End.Column - area.Start.Column
	} else {
		deletedColumns = area.End.Column
	}
	deletedRows := area.End.Row - area.Start.Row

	for _, cursor := range editor.Cursors {
		if cursor.Start.Row == area.End.Row && cursor.Start.Column > area.End.Column {
			cursor.Start.Column -= deletedColumns
		}
		if cursor.End.Row == area.End.Row && cursor.Start.Column >= area.End.Column {
			cursor.End.Column -= deletedColumns
		}
		if cursor.Start.Row >= area.End.Row {
			cursor.Start.Row -= deletedRows
			cursor.End.Row -= deletedRows
		}
	}
}

// Helper, prevents OOB
func slice(line []rune, start, end int) []rune {
	if end > len(line) || end < 0 {
		end = len(line)
	}
	if start > len(line) - 1 {
		return []rune{}
	}
	return line[start:end]
}

func (s *singleDelete) Undo(editor *Editor) {
	editor.Buffer.ChangeLine(s.area.Start.Row, s.originalLines[0])

	lineNumber := s.area.Start.Row + 1
	i := 1

	for i < len(s.originalLines) {
		editor.Buffer.AddLine(lineNumber, s.originalLines[i])
		lineNumber++
		i++
	}

	restoreCursors(editor, s.originalCursors)
}

func (s *singleDelete) Name() string {
	return "Single Delete"
}


type _delete struct {
	singleDeletes   []*singleDelete
	originalCursors []Range
}

func Delete() *_delete {
	return &_delete{}
}

func (d *_delete) Do(editor *Editor, cursor *Range, index int) {
	if index == len(editor.Cursors) - 1 {
		d.originalCursors = make([]Range, len(editor.Cursors))
		d.singleDeletes = make([]*singleDelete, len(editor.Cursors))
	}
	singleDelete := SingleDelete(*cursor)

	d.singleDeletes[index] = singleDelete
	d.originalCursors[index] = *cursor

	singleDelete.Do(editor)
	cursor.End.Row = cursor.Start.Row
	cursor.End.Column = cursor.Start.Column + 1
}

func (d *_delete) Undo(editor *Editor, cursor *Range, index int) {
	d.singleDeletes[index].Undo(editor)
	*cursor = d.originalCursors[index]
}

func (d *_delete) Name() string {
	return "Delete"
}

// By default, the cursor can be one more character to the right than there is
// in the line, representing the newline and allowing for insertions after the
// last character. This means the cursor end is OOB.
// This function simply moves those cursors ends into the start of the next line,
func cursorIncludeNewline(editor *Editor, cursor *Range) {
	if !isInBounds(editor, cursor.End) {
		// cursor is on last line
		if cursor.End.Row >= editor.Buffer.GetLength() - 1 {
			cursor.End.Column = ColumnSpan(editor, editor.Buffer.GetLine(cursor.End.Row)) + 1
		} else {
			cursor.End.Row++
			cursor.End.Column = 0
		}
	}
	if !isInBounds(editor, cursor.Start) {
		cursor.Start.Column = ColumnSpan(editor, editor.Buffer.GetLine(cursor.Start.Row))
	}
}

func isInBounds(editor *Editor, location Location) bool {
	if location.Column == 0 {
		return true
	}

	location.Column--
	line := editor.Buffer.GetLine(location.Row)
	if location.Column >= ColumnSpan(editor, line) {
		return false
	}
	return true
}
