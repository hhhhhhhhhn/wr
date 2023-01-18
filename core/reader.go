package core

import (
	"io"
	"unicode/utf8"
)

type EditorReader struct {
	editor        *Editor
	row           int
	index         int
	eof           bool
	currentLine   []rune
	pendingBytes  []byte
}

func (e *EditorReader) Read(out []byte) (read int, err error) {
	if e.eof {
		e.eof = false
		return 0, io.EOF
	}
	if e.pendingBytes != nil {
		if len(out) < len(e.pendingBytes) {
			return 0, io.ErrShortBuffer
		}
		copy(out, e.pendingBytes)
		read = len(e.pendingBytes)
		e.pendingBytes = nil
		return read, nil
	}
	for {
		char, length, err := e.ReadRune()
		if err != nil {
			copy(out, []byte{'\n'})
			read += 1
			e.eof = true
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
	if e.currentLine == nil {
		e.currentLine = e.editor.Buffer.GetLine(e.row)
	}

	if e.index >= len(e.currentLine) {
		if e.row < e.editor.Buffer.GetLength() - 1 {
			e.currentLine = nil
			e.row++
			e.index = 0
			return '\n', 1, nil
		} else {
			return 0, 0, io.EOF
		}
	}
	char = e.currentLine[e.index]
	length = utf8.RuneLen(char)
	e.index++
	return char, length, nil
}

func (e *EditorReader) UnreadRune() (char rune, length int, err error) {
	if e.index == 0 {
		if e.row == 0 {
			return 0, 0, io.EOF
		}
		e.row--
		e.index = len(e.editor.Buffer.GetLine(e.row))
		e.currentLine = nil
		return '\n', 1, nil
	}

	if e.currentLine == nil {
		e.currentLine = e.editor.Buffer.GetLine(e.row)
	}

	e.index--
	char = e.currentLine[e.index]
	length = utf8.RuneLen(char)
	return char, length, nil
}

func (e *EditorReader) SetLocation(row, col int) {
	e.eof = false
	e.row = row
	e.index = LocationToIndex(e.editor, Location{row, col})
	e.currentLine = nil
}

func (e *EditorReader) GetLocation() (row, col int) {
	return e.row, ColumnSpan(e.editor, e.editor.Buffer.GetLine(e.row)[:e.index])
}

func NewEditorReader(editor *Editor, row, col int) *EditorReader {
	reader := &EditorReader{editor: editor, eof: false}
	reader.SetLocation(row, col)
	return reader
}
