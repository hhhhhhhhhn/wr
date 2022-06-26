package main

import (
	"fmt"
	"strings"

	"github.com/hhhhhhhhhn/hexes"
	"github.com/hhhhhhhhhn/hexes/input"
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
	modeString = padWithSpaces(modeString, len(modeString), r.Cols)
	r.SetAttribute(attrStatus)
	r.SetString(row, 0, modeString)
	r.SetString(row, r.Cols-len(position), position)
}

func commandMode() {
	buffer := core.NewBuffer()
	buffer.Current = buffer.Current.Insert(0, [][]rune{{}})
	commandEditor := &core.Editor{Buffer: buffer}

	for {
		if len(commandEditor.Cursors) != 1 {
			core.SetCursors(0, 0, 0, 1)(commandEditor)
		}
		command := getLineAsString(commandEditor, 0)
		printCommand(renderer, command)
		event := getEvent()
		for event.EventType != input.KeyPressed {
			event = getEvent()
		}
		switch event.Chr {
		case input.ENTER:
			runCommand(command)
			return
		case input.ESCAPE:
			return
		case input.BACKSPACE:
			core.GoTo(core.Chars(-1))(commandEditor)
			core.SelectUntil(core.Chars(1))(commandEditor)
			core.AsEdit(core.Delete)(commandEditor)
			break
		default:
			core.AsEdit(core.Insert([]rune{event.Chr}))(commandEditor)
			break
		}
	}
}

func runCommand(command string) {
	printCommand(renderer, command)
}

func printCommand(r *hexes.Renderer, command string) {
	row := r.Rows - 1
	formatted := padWithSpaces(":" + command, len(command), r.Cols)

	r.SetAttribute(attrStatus)
	r.SetString(row, 0, formatted)
	r.SetAttribute(attrActive)
	r.Set(row, len(command)+1, ' ')

	out.Flush()
}
