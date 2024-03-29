package core

import "regexp"

type Edit func(*Editor)
type CursorEdit func(*Editor, *Cursor) // Only uses single cursor

// Used internally, for converting cursorEdits into Edits
func filter[T any](original []T, test func(T) bool) []T {
	filtered := make([]T, len(original))
	i := 0

	for _, element := range original {
		if test(element) {
			filtered[i] = element
			i++
		}
	}

	return filtered[:i]
}

func IsOOB(editor *Editor, cursor *Cursor) bool {
	lastRow := editor.Buffer.GetLength() - 1
	return !(cursor.Start.Row >= 0 &&
		   cursor.End.Row >= 0 &&
		   cursor.Start.Row <= lastRow &&
		   cursor.End.Row <= lastRow)
}

func removeOOBCursors(editor *Editor) {
	editor.Cursors = filter(editor.Cursors, func(cursor *Cursor) bool {
		return !IsOOB(editor, cursor)
	})
}

func AsEdit(cursorEdit CursorEdit) Edit {
	return func(editor *Editor) {
		sortedCursors := SortCursors(editor.Cursors)
		// CursorEdits are done in reverse cursor order, as, for example,
		// inserting a line messes with the location of the cursors
		// following the insert.
		for i := len(sortedCursors) - 1; i >= 0; i-- {
			cursorEdit(editor, sortedCursors[i])
		}

		removeOOBCursors(editor)
	}
}

type singleSplit struct {
	originalCursors []Range
	row             int
	column          int
}

func SingleSplit(row, column int) Edit {
	return func(editor *Editor) {
		line := editor.Buffer.GetLine(row)
		lineIndex := LocationToIndex(editor, Location{row, column})

		line1 := line[:lineIndex]
		line2 := line[lineIndex:]

		editor.Buffer.ChangeLine(row, line1)
		editor.Buffer.AddLine(row + 1, line2)

		for _, cursor := range editor.Cursors {
			if cursor.Start.Row > row {
				cursor.Start.Row++
				cursor.End.Row++
			} else if cursor.Start.Row == row && cursor.Start.Column >= column {
				cursor.Start.Column -= column
				cursor.End.Column -= column
				cursor.Start.Row++
				cursor.End.Row++
			}
		}
	}
}

func Split(editor *Editor, cursor *Cursor) {
	SingleSplit(cursor.Start.Row, cursor.Start.Column)(editor)
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

func SingleInsertInLine(insertion []rune, row, column int) Edit {
	return func(editor *Editor) {
		line := editor.Buffer.GetLine(row)
		lineIndex := LocationToIndex(editor, Location{row, column})
		newLine := Join(line[:lineIndex], insertion, line[lineIndex:])

		editor.Buffer.ChangeLine(row, newLine)
		insertedColumns := ColumnSpan(editor, insertion)

		for _, cursor := range editor.Cursors {
			if cursor.Start.Row == row && cursor.Start.Column >= column {
				cursor.Start.Column += insertedColumns
			}
			if cursor.End.Row == row && cursor.End.Column >= column {
				cursor.End.Column += insertedColumns
			}
		}
	}
}

func InsertInLine(insertion []rune) CursorEdit {
	return func(editor *Editor, cursor *Cursor) {
		SingleInsertInLine(insertion, cursor.Start.Row, cursor.Start.Column)(editor)
	}
}

func Insert(insertion []rune) CursorEdit {
	return func(editor *Editor, cursor *Cursor) {
		for i, line := range splitRune(insertion, '\n') {
			if i > 0 {
				Split(editor, cursor)
			}
			InsertInLine(line)(editor, cursor)
		}
	}
}

func SmartSplit(editor *Editor, cursor *Cursor) {
	indentation := GetIndentation(editor, cursor)
	Split(editor, cursor)
	InsertInLine(indentation)(editor, cursor)
}

var whitespace = regexp.MustCompile(`^\s*`)

func GetIndentation(editor *Editor, cursor *Cursor) []rune {
	reader := NewEditorReader(editor, cursor.Start.Row, 0)
	match := whitespace.FindReaderIndex(reader)
	if match == nil { // Shouldn't really happen, but safety first
		return []rune{}
	}
	reader.SetLocation(cursor.Start.Row, 0)
	indentation := []rune{}
	for i := 0; i < match[1]; i++ {
		chr, _, _ := reader.ReadRune()
		indentation = append(indentation, chr)
	}
	return indentation
}

func cursorInLineStart(cursor Cursor) Cursor {
	cursor.End.Row = cursor.Start.Row
	cursor.Start.Column = 0
	cursor.End.Column = 1
	return cursor
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

func SingleDelete(rangee Range) Edit {
	return func(editor *Editor) {
		cursorIncludeNewline(editor, &rangee)

		originalLines := [][]rune{editor.Buffer.GetLine(rangee.Start.Row)}

		for lineNumber := rangee.Start.Row + 1; lineNumber <= rangee.End.Row; lineNumber++ {
			originalLines = append(originalLines, editor.Buffer.GetLine(lineNumber))
		}

		cursorStartIndex := LocationToIndex(editor, rangee.Start)
		cursorEndIndex := LocationToIndex(editor, rangee.End)

		newLine := Join(
			slice(originalLines[0], 0, cursorStartIndex),
			slice(originalLines[len(originalLines)-1], cursorEndIndex, -1),
		)

		for lineNumber := rangee.End.Row; lineNumber >= rangee.Start.Row + 1; lineNumber-- {
			// NOTE: This is very inneficient, as all lines are relocated after
			// every delete
			editor.Buffer.RemoveLine(lineNumber)
		}
		editor.Buffer.ChangeLine(rangee.Start.Row, newLine)

		// The amount of deleted columns on the last line
		var deletedColumns int
		if rangee.Start.Row == rangee.End.Row {
			deletedColumns = rangee.End.Column - rangee.Start.Column
		} else {
			deletedColumns = rangee.End.Column
		}
		deletedRows := rangee.End.Row - rangee.Start.Row

		for _, cursor := range editor.Cursors {
			if cursor.Start.Row == rangee.End.Row && cursor.Start.Column > rangee.End.Column {
				cursor.Start.Column -= deletedColumns
			}
			if cursor.End.Row == rangee.End.Row && cursor.End.Column >= rangee.End.Column {
				cursor.End.Column -= deletedColumns
				if cursor.Start == cursor.End {
					cursor.End.Column++
				}
			}
			if cursor.Start.Row >= rangee.End.Row {
				cursor.Start.Row -= deletedRows
				cursor.End.Row -= deletedRows
			}
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

func Delete(editor *Editor, cursor *Cursor) {
	SingleDelete(cursor.Range)(editor)
}

func Yank(register int) CursorEdit {
	return func(editor *Editor, cursor *Cursor) {
		reader := NewEditorReader(editor, cursor.Start.Row, cursor.Start.Column)
		// FIXME: The localization is very inneficient
		yanked := []rune{}
		for {
			chr, _, err := reader.ReadRune()
			row, col := reader.GetLocation()
			if (row == cursor.End.Row && col > cursor.End.Column) ||
				row > cursor.End.Row ||
				err != nil {
					break
			}
			yanked = append(yanked, chr)
		}
		cursor.Registers[register] = yanked
	}
}

func Paste(register int) CursorEdit {
	return func (editor *Editor, cursor *Cursor) {
		Insert(cursor.Registers[register])(editor, cursor)
	}
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

func OnlyMainCursor(editor *Editor) {
	cursorsLength := len(editor.Cursors)
	for i := 0; i < cursorsLength-1; i++ {
		RemoveCursor(editor.Cursors[0])(editor)
	}
}

// Set cursors in StartRow, StartCol, EndRow, EndCol notation
func SetCursors(locations ...int) Edit {
	return func(editor *Editor) {
		for i := 0; i < len(locations)/4; i++ {
			PushCursor(&Cursor{Range: Range{
				Start: Location{locations[i*4], locations[i*4+1]},
				End: Location{locations[i*4+2], locations[i*4+3]},
			}})(editor)
		}
	}
}
