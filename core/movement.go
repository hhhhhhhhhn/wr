package core

type Movement func(*Editor, Range) Range

func Rows(rows int) Movement {
	return func(editor *Editor, cursor Range) Range {
		cursor.Start.Row += rows
		cursor.End.Row += rows
		return cursor
	}
}

func Columns(cols int) Movement {
	return func(editor *Editor, cursor Range) Range {
		cursor.Start.Column += cols
		cursor.End.Column += cols
		return cursor
	}
}

func Chars(chars int) Movement {
	return func(editor *Editor, cursor Range) Range {
		line := editor.Buffer.GetLine(cursor.Start.Row)
		cursorChrIndex := ColumnToIndex(editor, line, cursor.Start.Column)
		newCursorChrIndex := cursorChrIndex + chars

		if newCursorChrIndex < 0 && cursor.Start.Row == 0 {
			cursor.Start.Row--
			return cursor
		}

		// Go to the end of the previous line if on start
		if newCursorChrIndex < 0 {
			cursor.Start.Row--
			line = editor.Buffer.GetLine(cursor.Start.Row)
			cursor.Start.Column = ColumnSpan(editor, line)
			cursor.End.Row = cursor.Start.Row
			cursor.End.Column = cursor.Start.Column + 1

			// The +1 is because of the newline
			return Chars(newCursorChrIndex + 1)(editor, cursor)
		}

		// Go to start of next line if on end
		if newCursorChrIndex > len(line) {
			cursor.Start.Row++
			cursor.End.Row = cursor.Start.Row
			cursor.Start.Column = 0
			cursor.End.Column = 1

			if cursor.Start.Row < editor.Buffer.GetLength() {
				// The -1 is because of the newline
				return Chars(newCursorChrIndex - len(line) - 1)(editor, cursor)
			}
			return cursor
		}

		cursor.End.Row = cursor.Start.Row
		cursor.Start.Column = ColumnSpan(editor, line[:newCursorChrIndex])
		cursor.End.Column = cursor.Start.Column + 1
		return cursor
	}
}

type goTo struct {
	originalCursors []Range
	movement        Movement
	removeCursors   []*removeCursor
}

func GoTo(movement Movement) *goTo {
	return &goTo{movement: movement}
}

func (g *goTo) Do(e *Editor) {
	for _, cursor := range e.Cursors {
		g.originalCursors = append(g.originalCursors, *cursor)
		*cursor = g.movement(e, *cursor)
	}

	lastRow := e.Buffer.GetLength() - 1
	for _, cursor := range e.Cursors {
		if cursor.Start.Row < 0 ||
			cursor.End.Row < 0 ||
			cursor.Start.Row > lastRow ||
			cursor.End.Row > lastRow {
				removeCursor := RemoveCursor(cursor)
				removeCursor.Do(e)
				g.removeCursors = append(g.removeCursors, removeCursor)
			}
		if cursor.Start.Column < 0 {
			cursor.Start.Column = 0
		}
		if cursor.End.Column < 1 {
			cursor.End.Column = 1
		}
	}
}

func (g *goTo) Undo(e *Editor) {
	for i := len(g.removeCursors) - 1; i >= 0; i-- {
		g.removeCursors[i].Undo(e)
	}

	for i, cursor := range e.Cursors {
		*cursor = g.originalCursors[i]
	}

	g.removeCursors = nil
}

func (g *goTo) Name() string {
	return "Go To"
}

type selectUntil struct {
	originalCursors []Range
	movement        Movement
}

func SelectUntil(movement Movement) *selectUntil {
	return &selectUntil{movement: movement}
}

func (s *selectUntil) Do(e *Editor) {
	for _, cursor := range e.Cursors {
		s.originalCursors = append(s.originalCursors, *cursor)
		cursor.End = s.movement(e, *cursor).Start
	}

	lastRow := e.Buffer.GetLength() - 1
	for _, cursor := range e.Cursors {
		if cursor.End.Row < 0 ||
			cursor.End.Row > lastRow {
				cursor.End.Row = lastRow
				cursor.End.Column = ColumnSpan(e, e.Buffer.GetLine(lastRow))
			}
		if cursor.End.Column < 0 {
			cursor.End.Column = 0
		}
	}
}

func (s *selectUntil) Undo(e *Editor) {
	for i, cursor := range e.Cursors {
		*cursor = s.originalCursors[i]
	}
}

func (s *selectUntil) Name() string {
	return "Select Until"
}

type expandSelection struct {
	originalCursors []Range
	movement        Movement
}

func ExpandSelection(movement Movement) *expandSelection {
	return &expandSelection{movement: movement}
}

func (e *expandSelection) Do(editor *Editor) {
	for _, cursor := range editor.Cursors {
		e.originalCursors = append(e.originalCursors, *cursor)
		end := Range{Start: cursor.End, End: cursor.End}
		cursor.End = e.movement(editor, end).Start
	}

	lastRow := editor.Buffer.GetLength() - 1
	for _, cursor := range editor.Cursors {
		if cursor.End.Row < 0 ||
			cursor.End.Row > lastRow {
				cursor.End.Row = lastRow
				cursor.End.Column = ColumnSpan(editor, editor.Buffer.GetLine(lastRow))
			}
		if cursor.End.Column < 0 {
			cursor.End.Column = 0
		}
	}
}

func (e *expandSelection) Undo(editor *Editor) {
	for i, cursor := range editor.Cursors {
		*cursor = e.originalCursors[i]
	}
}

func (e *expandSelection) Name() string {
	return "Select Until"
}

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
