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

type pushCursorBelow struct {
	pushCursor *pushCursor
}

func PushCursorBelow() *pushCursorBelow {
	return &pushCursorBelow{}
}

func (p *pushCursorBelow) Do(editor *Editor) {
	if len(editor.Cursors) > 0 {
		lastCursor := editor.Cursors[len(editor.Cursors) - 1]
		if lastCursor.End.Row < editor.Buffer.GetLength() - 1 {
			rowOffset := lastCursor.End.Row - lastCursor.Start.Row + 1
			newCursor := &Range{
				Location{lastCursor.End.Row + rowOffset, lastCursor.Start.Column},
				Location{lastCursor.End.Row + rowOffset, lastCursor.End.Column},
			}
			p.pushCursor = PushCursor(newCursor)
			p.pushCursor.Do(editor)
		}
	}
}

func (p *pushCursorBelow) Undo(editor *Editor) {
	if p.pushCursor != nil {
		p.pushCursor.Undo(editor)
	}
}

func (p *pushCursorBelow) Name() string {
	return "Push Cursor Below"
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

type endMoveColumns struct {
	cols         int
	originalCursors map[*Range]Range
}

func EndMoveColumns(cols int) *endMoveColumns {
	return &endMoveColumns{cols, make(map[*Range]Range)}
}

func (e *endMoveColumns) Do(editor *Editor, cursor *Range) {
	e.originalCursors[cursor] = *cursor

	cursor.End.Column += e.cols

	if cursor.End.Column < 0 {
		cursor.End.Column = 0
	}
}

func (e *endMoveColumns) Undo(editor *Editor, cursor *Range) {
	*cursor = e.originalCursors[cursor]
}

func (e *endMoveColumns) Name() string {
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

type endMoveRows struct {
	rows            int
	originalCursors map[*Range]Range
}

func EndMoveRows(rows int) *endMoveRows {
	return &endMoveRows{rows, make(map[*Range]Range)}
}

func (e *endMoveRows) Do(editor *Editor, cursor *Range) {
	// NOTE: OOB is handled in wrapper
	e.originalCursors[cursor] = *cursor

	cursor.End.Row += e.rows
}

func (e *endMoveRows) Undo(editor *Editor, cursor *Range) {
	*cursor = e.originalCursors[cursor]
}

func (e *endMoveRows) Name() string {
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

	// The column in which each character/rune of the line is
	cursorChrIndex := ColumnToIndex(editor, line, cursor.Start.Column)
	newCursorChrIndex := cursorChrIndex + m.chars

	// OOB, new position is before start of file
	if newCursorChrIndex < 0 && cursor.Start.Row == 0 {
		cursor.Start.Row--
		return
	}

	// Go to the end of the previous line if on start
	if newCursorChrIndex < 0 {
		cursor.Start.Row--
		line = editor.Buffer.GetLine(cursor.Start.Row)
		cursor.Start.Column = ColumnSpan(editor, line)
		cursor.End.Row = cursor.Start.Row
		cursor.End.Column = cursor.Start.Column + 1
		return
	}

	// Go to start of next line if on end
	if newCursorChrIndex > len(line) {
		cursor.Start.Row++
		cursor.End.Row = cursor.Start.Row
		cursor.Start.Column = 0
		cursor.End.Column = 1
		return
	}

	cursor.End.Row = cursor.Start.Row
	cursor.Start.Column = ColumnSpan(editor, line[:newCursorChrIndex])
	cursor.End.Column = cursor.Start.Column + 1
}

func (m *moveChars) Undo(editor *Editor, cursor *Range) {
	*cursor = m.originalCursors[cursor]
}

func (m *moveChars) Name() string {
	return "Move Chars"
}

type endMoveChars struct {
	chars           int
	originalCursors map[*Range]Range
}

func EndMoveChars(chars int) *endMoveChars {
	return &endMoveChars{chars, make(map[*Range]Range)}
}

func (e *endMoveChars) Do(editor *Editor, cursor *Range) {
	e.originalCursors[cursor] = *cursor

	line := editor.Buffer.GetLine(cursor.Start.Row)

	// The column in which each character/rune of the line is
	cursorChrIndex := ColumnToIndex(editor, line, cursor.End.Column)
	newCursorChrIndex := cursorChrIndex + e.chars

	// OOB, new position is before start of file
	if newCursorChrIndex < 0 && cursor.End.Row == 0 {
		cursor.End.Row--
		return
	}

	// Go to the end of the previous line if on start
	if newCursorChrIndex < 0 {
		cursor.End.Row--
		line = editor.Buffer.GetLine(cursor.End.Row)
		cursor.End.Column = ColumnSpan(editor, line)
		return
	}

	// Go to start of next line if on end
	if newCursorChrIndex > len(line) {
		cursor.End.Row++
		cursor.End.Column = 1
		return
	}

	cursor.End.Column = ColumnSpan(editor, line[:newCursorChrIndex])
}

func (e *endMoveChars) Undo(editor *Editor, cursor *Range) {
	*cursor = e.originalCursors[cursor]
}

func (e *endMoveChars) Name() string {
	return "Move Chars"
}

type boundToLine struct {
	originalCursors map[*Range]Range
}

func BoundToLine() *boundToLine {
	return &boundToLine{originalCursors: make(map[*Range]Range)}
}

func (b *boundToLine) Do(editor *Editor, cursor *Range) {
	b.originalCursors[cursor] = *cursor

	startLine := editor.Buffer.GetLine(cursor.Start.Row)
	startColumnSpan := ColumnSpan(editor, startLine)

	var endColumnSpan int
	if cursor.Start.Row != cursor.End.Row {
		endLine := editor.Buffer.GetLine(cursor.End.Row)
		endColumnSpan = ColumnSpan(editor, endLine)
	} else {
		endColumnSpan = startColumnSpan
	}

	cursor.Start.Column = bound(cursor.Start.Column, 0, startColumnSpan)
	cursor.End.Column = bound(cursor.End.Column, 0, endColumnSpan + 1)
}

func (b *boundToLine) Undo(editor *Editor, cursor *Range) {
	*cursor = b.originalCursors[cursor]
}

func (b *boundToLine) Name() string {
	return "Bound To Line"
}

func bound(value, min, max int) int {
	if value < min {
		return min
	} else if value > max {
		return max
	}
	return value
}
