package main

import (
	"strings"
	"fmt"

	"github.com/hhhhhhhhhn/hexes"
	"github.com/hhhhhhhhhn/wr/core"
)

func handleScroll(e*core.Editor, renderRows int, currentScroll int) (newScroll int) {
	var lastCursorRow int
	if len(editor.Cursors) > 0 {
		lastCursorRow = e.Cursors[len(e.Cursors) - 1].Start.Row
	} else {
		lastCursorRow = 0
	}
	if lastCursorRow < currentScroll {
		currentScroll = lastCursorRow
	}
	if lastCursorRow > currentScroll + renderRows - 1 {
		currentScroll = lastCursorRow - renderRows + 1
	}
	return currentScroll
}

func padWithSpaces(str string, cols int, desiredCols int) string {
	if cols < desiredCols {
		str += strings.Repeat(" ", desiredCols - cols)
	}
	return str
}

func getLineAsString(e *core.Editor, row int) string {
	return strings.ReplaceAll(string(e.Buffer.GetLine(row)), "\t", strings.Repeat(" ", e.Config.Tabsize))
}

func printLine(e *core.Editor, r *hexes.Renderer, row int) {
	line := getLineAsString(e, row) + " "

	col := 0
	for _, chr := range line {
		withinCursor, withinLast := isWithinCursor(e, row, col)
		if withinCursor {
			if withinLast {
				r.SetAttribute(attrActive)
			} else {
				r.SetAttribute(attrCursor)
			}
		} else {
			r.SetAttribute(attrDefault)
		}
		r.SetString(row - scroll, col, string(chr))
		col += core.RuneWidth(e, chr)
	}

	r.SetAttribute(attrDefault)
	for col < r.Cols {
		r.Set(row - scroll, col, ' ')
		col += 1
	}
}

func PrintEditor(e *core.Editor, r *hexes.Renderer) {
	renderRows := r.Rows - 1 // Extra row for commands
	scroll = handleScroll(e, renderRows, scroll)

	lineAmount := e.Buffer.GetLength()

	var row int
	for row = scroll; row < scroll + renderRows && row < lineAmount; row++ {
		printLine(e, r, row)
	}

	for ;row < scroll + renderRows; row++ {
		r.SetString(row - scroll, 0, strings.Repeat(" ", r.Cols))
	}

	printStatusBar(e, r, statusString)

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

var modes = []string{}
var statusString string
var statusOk bool = true

func pushMode(mode string) {
	modes = append(modes, mode)
	updateStatusString()
}

func popMode() {
	modes = modes[:len(modes)-1]
	updateStatusString()
}

func updateStatusString() {
	statusString = strings.Join(modes, " > ")
	statusOk = true
}

func printStatusBar(e *core.Editor, r *hexes.Renderer, statusString string) {
	row := r.Rows - 1
	var position string
	if len(e.Cursors) > 0 {
		position = fmt.Sprintf("line %v, col %v ",
			editor.Cursors[len(editor.Cursors)-1].Start.Row,
			editor.Cursors[len(editor.Cursors)-1].Start.Column,
		)
	}

	statusString = " " + statusString
	statusString = padWithSpaces(statusString, len(statusString), r.Cols)
	if statusOk {
		r.SetAttribute(attrStatus)
	} else {
		r.SetAttribute(attrStatusErr)
	}
	r.SetString(row, 0, statusString)
	r.SetString(row, r.Cols-len(position), position)
}

