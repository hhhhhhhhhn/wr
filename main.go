package main

import (
	"github.com/hhhhhhhhhn/wr/core"
	"github.com/hhhhhhhhhn/hexes"
	"strings"
	"os"
	"time"
)

var scroll = 0

func main() {
	editor := core.Editor{Buffer: &core.Buffer{Lines: []string{"aaaa", "bbbb", "cccc"}}}
	renderer := hexes.New(os.Stdin, os.Stdout)
	renderer.Start()

	PrintEditor(&editor, renderer)
	time.Sleep(1000 * time.Millisecond)

	editor.Do(
		core.PushCursor(&core.Range{core.Location{1, 1}, core.Location{1, 3}}),
	)
	PrintEditor(&editor, renderer)
	time.Sleep(1000 * time.Millisecond)

	editor.CursorDo(core.Insert(
		"ASdasd",
	))
	time.Sleep(1000 * time.Millisecond)
	PrintEditor(&editor, renderer)

	editor.Do(
		core.PushCursor(&core.Range{core.Location{2, 1}, core.Location{2, 3}}),
	)
	time.Sleep(1000 * time.Millisecond)
	PrintEditor(&editor, renderer)

	editor.CursorDo(core.Delete())
	time.Sleep(1000 * time.Millisecond)
	PrintEditor(&editor, renderer)

	time.Sleep(1000 * time.Millisecond)

	renderer.End()
}

func PrintEditor(e *core.Editor, r *hexes.Renderer) {
	lineAmount := e.Buffer.GetLength()

	for row := scroll; row < scroll + r.Rows && row < lineAmount; row++ {
		line := strings.ReplaceAll(e.Buffer.GetLine(row), "\t", strings.Repeat(" ", e.Config.Tabsize)) + " "
		
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
		// Makes sure rest of line is empty
		for col < r.Cols {
			r.Set(row, col, " ")
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
