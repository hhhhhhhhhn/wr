package core

import (
	"io"
)

type EditorReader struct {
	editor  *Editor
	row     int
	col     int
	pending []byte
}

func (e *EditorReader) Read(out []byte) (read int, err error) {
	if e.row >= e.editor.Buffer.GetLength() || e.row < 0 {
		return 0, io.EOF
	}
	var buffer []byte

	if e.pending != nil {
		buffer = e.pending
		e.pending = nil
	} else {
		line := e.editor.Buffer.GetLine(e.row)
		index := ColumnToIndex(e.editor, line, e.col)
		buffer = []byte(string(line[index:]) + "\n")
	}

	if len(out) < len(buffer) {
		e.pending = buffer[len(out):]
		read = len(out)
	} else {
		read = len(buffer)
		e.row++
		e.col = 0
	}

	copy(out, buffer)
	return read, nil
}

func (e *EditorReader) GoTo(row, col int) {
	e.row = row
	e.col = col
	e.pending = nil
}

func (e *EditorReader) GetLocation() (row, col int) {
	return e.row, e.col
}

func NewEditorReader(editor *Editor, row, col int) *EditorReader {
	return &EditorReader{editor: editor, row: row, col: col}
}
