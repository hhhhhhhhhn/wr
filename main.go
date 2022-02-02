package main

import (
	"bufio"
	"os"
	"strings"
	"unicode"

	"github.com/hhhhhhhhhn/hexes"
	"github.com/hhhhhhhhhn/wr/core"
)

var scroll = 0

func main() {
	editor := core.Editor{Buffer: &core.Buffer{Lines: []string{"aaaa", "bbbb", "cccc"}}}
	renderer := hexes.New(os.Stdin, os.Stdout)
	renderer.Start()
	editor.Do(
		core.PushCursor(&core.Range{core.Location{0, 0}, core.Location{0, 1}}),
	)
	editor.Do(core.UndoMarker())

	in := bufio.NewReader(os.Stdin)

	PrintEditor(&editor, renderer)
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
			editor.CursorDo(core.MoveRows(1))
			break
		case 'k':
			editor.CursorDo(core.MoveRows(-1))
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
		default:
			if unicode.IsGraphic(chr) {
				editor.CursorDo(core.Insert(string(chr)))
			}
			break
		}
		PrintEditor(&editor, renderer)
	}
}

func PrintEditor(e *core.Editor, r *hexes.Renderer) {
	lineAmount := e.Buffer.GetLength()

	for row := scroll; row < scroll + r.Rows && row < lineAmount; row++ {
		line := strings.ReplaceAll(e.Buffer.GetLine(row), "\t", strings.Repeat(" ", e.Config.Tabsize))
		line += strings.Repeat(" ", r.Cols - len(line))
		
		col := 0
		for _, chr := range line {
			if isWithinCursor(e, row, col) {
				r.SetAttribute(hexes.REVERSE)
			} else {
				r.SetAttribute(r.DefaultAttribute)
			}
			r.SetString(row, col, string(chr))
			col++
		}
	}
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
