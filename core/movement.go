package core

import (
	"regexp"
)

var OOBCursor = Cursor{Range: Range{Start: Location{-1, -1}}}

type Movement func(*Editor, Cursor) Cursor

func Rows(rows int) Movement {
	return func(editor *Editor, cursor Cursor) Cursor {
		cursor.Start.Row += rows
		cursor.End.Row += rows
		return cursor
	}
}

func Columns(cols int) Movement {
	return func(editor *Editor, cursor Cursor) Cursor {
		cursor.Start.Column += cols
		cursor.End.Column += cols
		return cursor
	}
}

func EndOfLine(editor *Editor, cursor Cursor) Cursor {
	line := editor.Buffer.GetLine(cursor.Start.Row)
	cursor.Start.Column = ColumnSpan(editor, line)
	cursor.End.Column = cursor.Start.Column+1
	cursor.End.Row = cursor.Start.Row
	return cursor
}

func StartOfLine (editor *Editor, cursor Cursor) Cursor {
	cursor.Start.Column = 0
	cursor.End.Column = 1
	cursor.End.Row = cursor.Start.Row
	return cursor
}

func Unselect(_ *Editor, cursor Cursor) Cursor {
	cursor.End.Row = cursor.Start.Row
	cursor.End.Column = cursor.Start.Column + 1
	return cursor
}

func Chars(chars int) Movement {
	return func(editor *Editor, cursor Cursor) Cursor {
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

// NOTE: RegexCursor must have a "Cursor" group for this to work, and start with "^"
func RegexCursor(regex *regexp.Regexp, times int) Movement {
	if times < 0 {
		times--
	}
	return func(editor *Editor, cursor Cursor) Cursor {
		reader := NewEditorReader(editor, cursor.Start.Row, cursor.Start.Column)
		timesLeft := abs(times)
		eof := false
		for timesLeft > 0 {
			var err error
			if times > 0 {
				_, _, err = reader.ReadRune()
			} else {
				_, _, err = reader.UnreadRune()
			}

			row, col := reader.GetLocation()

			if row == -1 || err != nil{
				eof = true
				break
			}


			found := regex.MatchReader(reader)
			reader.SetLocation(row, col)

			if found {
				timesLeft--
			}
		}
		for i := 0; !eof && i < regex.SubexpIndex("Cursor"); i++ {
			reader.ReadRune()
		}
		row, col := reader.GetLocation()
		cursor.Start.Row = row
		cursor.End.Row = row
		cursor.Start.Column = col
		cursor.End.Column = col+1
		return cursor
	}
}

// NOTE: Regex must start with "^"
func Regex(regex *regexp.Regexp, times int) Movement {
	return func(editor *Editor, cursor Cursor) Cursor {
		reader := NewEditorReader(editor, cursor.Start.Row, cursor.Start.Column)
		timesLeft := abs(times)
		var endRow, endCol int

		for timesLeft > 0 {
			var err error
			if times > 0 {
				_, _, err = reader.ReadRune()
			} else {
				_, _, err = reader.UnreadRune()
			}

			row, col := reader.GetLocation()

			if row == -1 || err != nil{
				break
			}

			found := regex.MatchReader(reader)
			reader.UnreadRune()
			reader.UnreadRune()
			endRow, endCol = reader.GetLocation()

			reader.SetLocation(row, col)

			if found {
				timesLeft--
			}
		}
		row, col := reader.GetLocation()
		cursor.Start.Row = row
		cursor.End.Row = endRow
		cursor.Start.Column = col
		cursor.End.Column = endCol
		if comesFirst(cursor.End, cursor.Start) {
			cursor.End = cursor.Start
			cursor.End.Column += 1
		}
		return cursor
	}
}

func abs(a int) int {
	if a > 0 {
		return a
	}
	return -a
}

func Words(words int) Movement {
	return RegexCursor(regexp.MustCompile(`^\s(?P<Cursor>)\S`), words)
}

func GoTo(movement Movement) Edit {
	return func(editor *Editor) {
		for _, cursor := range editor.Cursors {
			*cursor = movement(editor, *cursor)
		}

		removeOOBCursors(editor)

		for _, cursor := range editor.Cursors {
			if cursor.Start.Column < 0 {
				cursor.Start.Column = 0
			}
			if cursor.End.Column < 1 {
				cursor.End.Column = 1
			}
		}
	}
}

type selectUntil struct {
	originalCursors []Range
	movement        Movement
}


func SelectUntil(movement Movement) Edit {
	return func(editor *Editor) {
		for _, cursor := range editor.Cursors {
			cursor.End = movement(editor, *cursor).Start
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
}

type expandSelection struct {
	originalCursors []Range
	movement        Movement
}

func ExpandSelection(movement Movement) Edit {
	return func(editor *Editor) {
		for _, cursor := range editor.Cursors {
			end := Range{Start: cursor.End, End: cursor.End}
			cursor.End = movement(editor, Cursor{Range: end}).Start
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
}

func PushCursor(cursor *Cursor) Edit {
	return func(editor *Editor) {
		editor.Cursors = append(editor.Cursors, cursor)
	}
}

func RemoveCursor(removedCursor *Cursor) Edit {
	return func(editor *Editor) {
		editor.Cursors = filter(editor.Cursors, func(cursor *Cursor) bool {
			return cursor != removedCursor
		})
	}
}

// If new position overlaps, removes last
func PushCursorFromLast(movement Movement) Edit {
	return func(editor *Editor) {
		if len(editor.Cursors) > 0 {
			lastCursor := editor.Cursors[len(editor.Cursors) - 1]
			newCursor := movement(editor, *lastCursor)
			if isWithinACursor(editor, &newCursor) {
				RemoveCursor(lastCursor)(editor)
			} else {
				PushCursor(&newCursor)(editor)
				// TODO: Check only last cursor
				removeOOBCursors(editor)
			}
		}
	}
}

func isWithinACursor(editor *Editor, cursor *Cursor) bool {
	row, col := cursor.Start.Row, cursor.Start.Column
	for _, cursor := range editor.Cursors {
		if (
			((row == cursor.Start.Row && col >= cursor.Start.Column) || (row > cursor.Start.Row)) &&
			((row == cursor.End.Row && col < cursor.End.Column) || (row < cursor.End.Row))){
				return true
			}
	}
	return false
}

func comesFirst(a Location, b Location) bool {
	return !(a.Row > b.Row || (a.Row == b.Row && a.Column >= b.Column))
}
