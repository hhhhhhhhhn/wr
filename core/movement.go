package core

type pushCursor struct {
	_range *Range
}

func PushCursor(_range *Range) *pushCursor {
	return &pushCursor{_range}
}

func (p *pushCursor) Do(e *Editor) {
	e.Cursors = append(e.Cursors, p._range)
}

func (p *pushCursor) Undo(e *Editor) {
	e.Cursors = e.Cursors[:len(e.Cursors) - 1]
}

func (p *pushCursor) Name() string {
	return "Push Cursor"
}

type removeCursor struct {
	removedCursor   *Range
	originalCursors []*Range
}

func RemoveCursor(cursor *Range) *removeCursor {
	return &removeCursor{removedCursor: cursor, originalCursors: []*Range{}}
}

func (r *removeCursor) Do(e *Editor) {
	newCursors := []*Range{}
	for _, cursor := range e.Cursors {
		if cursor != r.removedCursor {
			newCursors = append(newCursors, cursor)
		}
		r.originalCursors = append(r.originalCursors, cursor)
	}
	e.Cursors = newCursors
}

func (r *removeCursor) Undo(e *Editor) {
	e.Cursors = r.originalCursors
}

func (r *removeCursor) Name() string {
	return "Remove Cursor"
}


type moveColumns struct {
	cols         int
	originalCursors map[*Range]Range
}

func MoveColumns(cols int) *moveColumns {
	return &moveColumns{cols, make(map[*Range]Range)}
}

func (m *moveColumns) Do(editor *Editor, cursor *Range) {
	m.originalCursors[cursor] = *cursor

	cursor.Start.Column += m.cols

	if cursor.Start.Column < 0 {
		cursor.Start.Column = 0
	}

	cursor.End.Row = cursor.Start.Row
	cursor.End.Column = cursor.Start.Column + 1
}

func (m *moveColumns) Undo(editor *Editor, cursor *Range) {
	*cursor = m.originalCursors[cursor]
}

func (m *moveColumns) Name() string {
	return "Move Columns"
}

type moveRows struct {
	rows            int
	originalCursors map[*Range]Range
}

func MoveRows(rows int) *moveRows {
	return &moveRows{rows, make(map[*Range]Range)}
}

func (m *moveRows) Do(editor *Editor, cursor *Range) {
	// NOTE: OOB is handled in wrapper
	m.originalCursors[cursor] = *cursor

	cursor.Start.Row += m.rows
	cursor.End.Row = cursor.Start.Row
	cursor.End.Column = cursor.Start.Column + 1
}

func (m *moveRows) Undo(editor *Editor, cursor *Range) {
	*cursor = m.originalCursors[cursor]
}

func (m *moveRows) Name() string {
	return "Move Rows"
}

type moveChars struct {
	chars           int
	originalCursors map[*Range]Range
}

func MoveChars(chars int) *moveChars {
	return &moveChars{chars, make(map[*Range]Range)}
}

func (m *moveChars) Do(editor *Editor, cursor *Range) {
	m.originalCursors[cursor] = *cursor

	line := editor.Buffer.GetLine(cursor.Start.Row)
	lineIndex := LocationToLineIndex(editor, cursor.Start)
	lineIndex += m.chars

	if lineIndex < 0 {
		cursor.Start.Row--
		// Prevents crash
		if cursor.Start.Row >= 0 {
			line = editor.Buffer.GetLine(cursor.Start.Row)
		} else {
			line = ""
		}
		col := StringColumnSpan(editor, line)
		cursor.Start.Column = col
		cursor.End.Row = cursor.Start.Row
		cursor.End.Column = col + 1
		return
	}

	if lineIndex >= len(line) + 1 {
		cursor.Start.Row++
		cursor.End.Row = cursor.Start.Row
		cursor.Start.Column = 0
		cursor.End.Column = 1
		return
	}

	col := LineIndexToColumn(editor, lineIndex, line)

	cursor.End.Row = cursor.Start.Row
	cursor.Start.Column = col
	cursor.End.Column = col + 1
}

func (m *moveChars) Undo(editor *Editor, cursor *Range) {
	*cursor = m.originalCursors[cursor]
}

func (m *moveChars) Name() string {
	return "Move Chars"
}
