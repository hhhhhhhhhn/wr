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

	// The column in which each character/rune of the line is
	cols  := []int{0}

	for _, chr := range line {
		cols = append(cols, cols[len(cols)-1] + RuneWidth(editor, chr))
	}

	cursorChrIndex := -1
	for i, chrCol := range cols {
		if chrCol > cursor.Start.Column {
			cursorChrIndex = i-1
			break
		}
	}
	if cursorChrIndex == -1 {
		cursorChrIndex = len(cols) - 1
	}

	newCursorChrIndex := cursorChrIndex + m.chars

	// Go to the end of the previous line
	if newCursorChrIndex < 0 {
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


	if newCursorChrIndex >= len(cols) {
		cursor.Start.Row++
		cursor.End.Row = cursor.Start.Row
		cursor.Start.Column = 0
		cursor.End.Column = 1
		return
	}

	cursor.End.Row = cursor.Start.Row
	cursor.Start.Column = cols[newCursorChrIndex]
	cursor.End.Column = cursor.Start.Column + 1
}

func (m *moveChars) Undo(editor *Editor, cursor *Range) {
	*cursor = m.originalCursors[cursor]
}

func (m *moveChars) Name() string {
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
	startColumnSpan := StringColumnSpan(editor, startLine)

	var endColumnSpan int
	if cursor.Start.Row != cursor.End.Row {
		endLine := editor.Buffer.GetLine(cursor.End.Row)
		endColumnSpan = StringColumnSpan(editor, endLine)
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
