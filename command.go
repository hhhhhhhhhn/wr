package main

import (
	"fmt"
	"strings"

	"github.com/hhhhhhhhhn/hexes"
	"github.com/hhhhhhhhhn/wr/core"
)

var modes = []string{}
var modeString string

func pushMode(mode string) {
	modes = append(modes, mode)
	updateModeString()
}

func popMode() {
	modes = modes[:len(modes)-1]
	updateModeString()
}

func updateModeString() {
	modeString = strings.Join(modes, " > ")
}

func printStatusBar(e *core.Editor, r *hexes.Renderer, modeString string) {
	row := r.Rows - 1
	var position string
	if len(e.Cursors) > 0 {
		position = fmt.Sprintf("line %v, col %v ",
			editor.Cursors[len(editor.Cursors)-1].Start.Row,
			editor.Cursors[len(editor.Cursors)-1].Start.Column,
		)
	}

	modeString = " " + modeString
	if len(modeString) < r.Cols {
		modeString += strings.Repeat(" ", r.Cols - len(modeString))
	}
	r.SetAttribute(hexes.REVERSE)
	r.SetString(row, 0, modeString)
	r.SetString(row, r.Cols-len(position), position)
}

func commandMode() {
}
