package main

import (
	"strings"
	"github.com/hhhhhhhhhn/hexes"
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

func printStatusBar(r *hexes.Renderer, modeString string) {
	row := r.Rows - 1
	if len(modeString) < r.Cols {
		modeString += strings.Repeat(" ", r.Cols - len(modeString))
	}
	r.SetAttribute(hexes.REVERSE)
	r.SetString(row, 0, modeString)
}

func commandMode() {
}
