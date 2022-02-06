package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/hhhhhhhhhn/hexes"
	"github.com/hhhhhhhhhn/wr/core"
)

var scroll = 0
var out = bufio.NewWriterSize(os.Stdout, 4096)
var editor core.Editor
var in     *bufio.Reader
var renderer *hexes.Renderer

func normalMode() {
	for {
		chr, _, _ := in.ReadRune()
		switch(chr) {
		case 'H':
			editor.CursorDo(core.MoveColumns(-1))
			break
		case 'L':
			editor.CursorDo(core.MoveColumns(1))
			break
		case 'h':
			editor.CursorDo(core.MoveChars(-1))
			break
		case 'l':
			editor.CursorDo(core.MoveChars(1))
			break
		case 'j':
			editor.CursorDo(core.MoveRows(1))
			break
		case 'k':
			editor.CursorDo(core.MoveRows(-1))
			break
		case 'J':
			scroll++
			break
		case 'K':
			if scroll > 0 {
				scroll--
			}
			break
		case 'u':
			editor.Do(core.UndoMarker())
			break
		case 'U':
			editor.Undo()
			break
		case 'r':
			editor.Redo()
			break
		case 'R':
			renderer.Refresh()
			break
		case 'd': // Backspace
			editor.CursorDo(core.Delete())
			break
		case 'V':
			multiCursorMode()
		case 'i':
			insertMode()
			break
		default:
			break
		}
		PrintEditor(&editor, renderer)
	}
}

func insertMode() {
	editor.Do(core.UndoMarker())
	for {
		chr, _, _ := in.ReadRune()
		switch(chr) {
		case 'q':
			return
		case 127: // Backspace
			editor.CursorDo(core.MoveChars(-1))
			editor.CursorDo(core.Delete())
			break
		default:
			if unicode.IsGraphic(chr) || chr == '\t' || chr == '\n' {
				editor.CursorDo(core.Insert(string(chr)))
			} else {
				editor.CursorDo(core.Insert(fmt.Sprint(chr)))
			}
			break
		}
		PrintEditor(&editor, renderer)
	}
}

func multiCursorMode() {
	for {
		chr, _, _ := in.ReadRune()
		switch(chr) {
		case 'h':
			editor.CursorDo(core.MoveColumns(-1))
			break
		case 'l':
			editor.CursorDo(core.MoveColumns(1))
			break
		case 'j':
			editor.Do(core.PushCursorBelow())
			break
		case 'k':
			if len(editor.Cursors) > 1 {
				editor.Do(core.RemoveCursor(editor.Cursors[len(editor.Cursors) - 1]))
			}
			break
		case 'J':
			scroll++
			break
		case 'K':
			if scroll > 0 {
				scroll--
			}
			break
		case 'u':
			editor.Do(core.UndoMarker())
			break
		case 'U':
			editor.Undo()
			break
		case 'r':
			editor.Redo()
			break
		case 'R':
			renderer.Refresh()
			break
		case 'd':
			editor.CursorDo(core.Delete())
			break
		case 'q':
			return
		case 'i':
			insertMode()
			break
		default:
			break
		}
		PrintEditor(&editor, renderer)
	}
}

func main() {
	editor = core.Editor{Buffer: &core.Buffer{Lines: []string{"aaaa", "bbbb", "cccc"}}, Config: core.EditorConfig{Tabsize: 4}}
	renderer = hexes.New(os.Stdin, out)
	renderer.Start()
	editor.Do(
		core.PushCursor(&core.Range{core.Location{0, 0}, core.Location{0, 1}}),
	)

	in = bufio.NewReader(os.Stdin)

	PrintEditor(&editor, renderer)
	normalMode()
}

func PrintEditor(e *core.Editor, r *hexes.Renderer) {
	lineAmount := e.Buffer.GetLength()

	var row int
	for row = scroll; row < scroll + r.Rows && row < lineAmount; row++ {
		line := strings.ReplaceAll(e.Buffer.GetLine(row), "\t", strings.Repeat(" ", e.Config.Tabsize))
		line += strings.Repeat(" ", r.Cols - core.StringColumnSpan(e, line))

		col := 0
		for _, chr := range line {
			if isWithinCursor(e, row, col) {
				r.SetAttribute(hexes.REVERSE)
			} else {
				r.SetAttribute(r.DefaultAttribute)
			}
			r.SetString(row - scroll, col, string(chr))
			col += core.RuneWidth(e, chr)
		}
	}

	for ;row < scroll + r.Rows; row++ {
		r.SetString(row - scroll, 0, strings.Repeat(" ", r.Cols))
	}

	out.Flush()
}

func isWithinCursor(e *core.Editor, row, col int) bool {
	for _, cursor := range e.Cursors {
		if (row >= cursor.Start.Row &&
			col >= cursor.Start.Column &&
			row <= cursor.End.Row &&
			col < cursor.End.Column) {
				return true
			}
	}
	return false
}
