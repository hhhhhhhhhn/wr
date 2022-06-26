package main

import (
	"strings"

	"github.com/hhhhhhhhhn/hexes"
	"github.com/hhhhhhhhhn/wr/core"
)


func PrintEditor(e *core.Editor, r *hexes.Renderer) {
	renderRows := r.Rows - 1 // Extra row for commands

	var lastCursorRow int
	if len(editor.Cursors) > 0 {
		lastCursorRow = e.Cursors[len(e.Cursors) - 1].Start.Row
	} else {
		lastCursorRow = 0
	}
	if lastCursorRow < scroll {
		scroll = lastCursorRow
	}
	if lastCursorRow > scroll + renderRows - 1 {
		scroll = lastCursorRow - renderRows + 1
	}

	lineAmount := e.Buffer.GetLength()

	var row int
	for row = scroll; row < scroll + renderRows && row < lineAmount; row++ {
		line := strings.ReplaceAll(string(e.Buffer.GetLine(row)), "\t", strings.Repeat(" ", e.Config.Tabsize))
		columnSpan := core.ColumnSpan(e, []rune(line))

		if columnSpan < r.Cols {
			line += strings.Repeat(" ", r.Cols - columnSpan)
		}

		col := 0
		for _, chr := range line {
			withinCursor, withinLast := isWithinCursor(e, row, col)
			if withinCursor {
				if withinLast {
					r.SetAttribute(hexes.MAGENTA + hexes.REVERSE)
				} else {
					r.SetAttribute(hexes.REVERSE)
				}
			} else {
				r.SetAttribute(r.DefaultAttribute)
			}
			r.SetString(row - scroll, col, string(chr))
			col += core.RuneWidth(e, chr)
		}
	}

	for ;row < scroll + renderRows; row++ {
		r.SetString(row - scroll, 0, strings.Repeat(" ", r.Cols))
	}

	out.Flush()
}

func isWithinCursor(e *core.Editor, row, col int) (isWithin bool, isLast bool) {
	var cursors []*core.Cursor
	if len(e.Cursors) > 25 {
		cursors = e.Cursors[len(e.Cursors)-25:]
	} else {
		cursors = e.Cursors
	}
	for i, cursor := range cursors {
		if (
			((row == cursor.Start.Row && col >= cursor.Start.Column) || (row > cursor.Start.Row)) &&
			((row == cursor.End.Row && col < cursor.End.Column) || (row < cursor.End.Row))){
				if i == len(cursors) - 1 {
					return true, true
				}
				return true, false
			}
	}
	return false, false
}