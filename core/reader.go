package core

import (
	"io"
	"unicode/utf8"
)

type EditorReader struct {
	editor        *Editor
	row           int
	index         int
	remainingLine []rune
	pendingBytes  []byte

}

func (e *EditorReader) Read(out []byte) (read int, err error) {
	if e.pendingBytes != nil {
		if len(out) < len(e.pendingBytes) {
			return 0, io.ErrShortBuffer
		}
		copy(out, e.pendingBytes)
		read = len(e.pendingBytes)
		e.pendingBytes = nil
		return read, nil
	}
	if e.row >= e.editor.Buffer.GetLength() || e.row < 0 {
		return 0, io.EOF
	}
	for {
		char, length, err := e.ReadRune()
		if err != nil {
			break
		}
		if len(out) < length {
			e.pendingBytes = make([]byte, length)
			utf8.EncodeRune(e.pendingBytes, char)
			copy(out, e.pendingBytes)
			e.pendingBytes = e.pendingBytes[len(out):]
			read += len(out)
			break
		} else {
			utf8.EncodeRune(out, char)
			out = out[length:]
			read += length
		}
	}
	return read, nil
}

func (e *EditorReader) ReadRune() (char rune, length int, err error) {
	if e.row >= e.editor.Buffer.GetLength() || e.row < 0 {
		return 0, 0, io.EOF
	}

	if e.remainingLine == nil {
		e.remainingLine = e.editor.Buffer.GetLine(e.row)[e.index:]
	}

	if len(e.remainingLine) == 0 {
		e.remainingLine = nil
		e.row++
		e.index = 0
		return '\n', 1, nil
	}
	char = e.remainingLine[0]
	length = utf8.RuneLen(char)
	e.remainingLine = e.remainingLine[1:]
	e.index++
	return char, length, nil
}

func (e *EditorReader) UnreadRune() (err error) {
	if e.index == 0 {
		if e.row == 0 {
			return io.EOF
		}
		e.row--
		e.index = len(e.editor.Buffer.GetLine(e.row))
	} else {
		e.index--
	}
	return nil
}

func (e *EditorReader) SetLocation(row, col int) {
	e.row = row
	e.index = LocationToIndex(e.editor, Location{row, col})
	e.remainingLine = nil
}

func (e *EditorReader) GetLocation() (row, col int) {
	if e.row >= e.editor.Buffer.GetLength() {
		return -1, -1
	}
	return e.row, ColumnSpan(e.editor, e.editor.Buffer.GetLine(e.row)[:e.index])
}

func NewEditorReader(editor *Editor, row, col int) *EditorReader {
	reader := &EditorReader{editor: editor}
	reader.SetLocation(row, col)
	return reader
}
