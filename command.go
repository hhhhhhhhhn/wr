package main

import (
	"regexp"
	"strings"

	"github.com/hhhhhhhhhn/hexes"
	"github.com/hhhhhhhhhn/hexes/input"
	"github.com/hhhhhhhhhn/wr/core"
)

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
			out, _ := runCommand(command)
			modeString = out
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

func printCommand(r *hexes.Renderer, command string) {
	row := r.Rows - 1
	formatted := padWithSpaces(":" + command, len(command), r.Cols)

	r.SetAttribute(attrStatus)
	r.SetString(row, 0, formatted)
	r.SetAttribute(attrActive)
	r.Set(row, len(command)+1, ' ')

	out.Flush()
}

func runCommand(command string) (output string, ok bool) {
	args := strings.Split(command, " ")
	if len(args) == 0 {
		return
	}
	function, ok := commands[args[0]]
	if !ok && len(args[0]) > 0 {
		function, ok = commands[string(args[0][0])]
	}
	if !ok {
		return "command not found", false
	}
	return function(args)
}

var commands = map[string] func([]string)(output string, ok bool) {
	"w": func(args []string) (string, bool) {
		if len(args) > 1 {
			editor.Global["Filename"] = strings.Join(args[1:], " ")
		}
		err := core.SaveToFile(editor, editor.Global["Filename"].(string))
		if err != nil {
			return err.Error(), false
		}
		return "saved " + editor.Global["Filename"].(string), true
	},
	"q": func([]string) (string, bool) {
		quit()
		return "", true
	},
	"wq": func([]string) (string, bool) {
		err := core.SaveToFile(editor, "file.txt")
		if err != nil {
			return err.Error(), false
		}
		quit()
		return "", true
	},
	"hello": func([]string) (string, bool) {
		return "there",  true
	},
	"/": func(args []string) (string, bool) {
		regexString := "^" + strings.Join(args, " ")[1:]
		regex, err := regexp.Compile(regexString)
		if err != nil {
			return err.Error(), false
		}
		editor.Global["Regex"] = regex
		return "", true
	},
}
