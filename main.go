package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/hhhhhhhhhn/hexes"
	"github.com/hhhhhhhhhn/hexes/input"
	"github.com/hhhhhhhhhn/wr/core"
)

var scroll = 0
var out = bufio.NewWriterSize(os.Stdout, 4096)
var editor core.Editor
var renderer *hexes.Renderer
var listener  *input.Listener

func normalMode() {
	multicursor := false
	visual      := false
	for {
		event := listener.GetEvent()
		if event.EventType != input.KeyPressed {
			continue
		}
		switch(event.Chr) {
		case 'H':
			if visual {
				editor.CursorDo(core.EndMoveColumns(-1))
			} else {
				editor.CursorDo(core.MoveColumns(-1))
			}
			break
		case 'L':
			if visual {
				editor.CursorDo(core.EndMoveColumns(1))
			} else {
				editor.CursorDo(core.MoveColumns(1))
			}
			break
		case 'h':
			if visual {
				editor.CursorDo(core.EndMoveColumns(-1))
			} else {
				editor.CursorDo(core.MoveChars(-1))
			}
			break
		case 'l':
			if visual {
				editor.CursorDo(core.EndMoveColumns(1))
			} else {
				editor.CursorDo(core.MoveChars(1))
			}
			break
		case 'j':
			if multicursor {
				editor.Do(core.PushCursorBelow())
			} else {
				editor.CursorDo(core.MoveRows(1))
			}
			break
		case 'k':
			if multicursor {
				if len(editor.Cursors) > 1 {
					editor.Do(core.RemoveCursor(editor.Cursors[len(editor.Cursors) - 1]))
				}
			} else {
				editor.CursorDo(core.MoveRows(-1))
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
			visual = false
			break
		case 22: // <C-v>
			multicursor = true
		case 'v':
			visual = true
		case 'i':
			insertMode()
			visual = false
			break
		case input.ESCAPE:
			visual = false
			multicursor = false
			for i := 0; i < len(editor.Cursors)-1; i++ {
				editor.Do(core.RemoveCursor(editor.Cursors[0]))
			}
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
		event := listener.GetEvent()
		if event.EventType != input.KeyPressed {
			continue
		}
		switch(event.Chr) {
		case input.ESCAPE:
			return
		case input.BACKSPACE:
			editor.CursorDo(core.MoveChars(-1))
			editor.CursorDo(core.Delete())
			break
		default:
			if unicode.IsGraphic(event.Chr) || event.Chr == '\t' || event.Chr == '\n' {
				editor.CursorDo(core.Insert(string(event.Chr)))
			} else {
				editor.CursorDo(core.Insert(fmt.Sprint(event.Chr)))
			}
			break
		}
		PrintEditor(&editor, renderer)
	}
}

func main() {
	editor = core.Editor{Buffer: &core.Buffer{Lines: []string{"aaaa", "bbbb", "cccc"}}, Config: core.EditorConfig{Tabsize: 4}}
	renderer = hexes.New(os.Stdin, out)
	listener = input.New(os.Stderr)
	renderer.Start()
	editor.Do(
		core.PushCursor(&core.Range{core.Location{0, 0}, core.Location{0, 1}}),
	)

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
