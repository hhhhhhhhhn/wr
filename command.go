package main

import (
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"

	"github.com/hhhhhhhhhn/hexes/input"
	"github.com/hhhhhhhhhn/wr/core"
	"github.com/hhhhhhhhhn/wr/treesitter"
)

func commandMode(command string) {
	buffer := core.NewBuffer()
	buffer.Current = buffer.Current.Insert(0, [][]rune{[]rune(command)})
	commandEditor := &core.Editor{Buffer: buffer}
	col := core.ColumnSpan(commandEditor, buffer.GetLine(0))

	for {
		if len(commandEditor.Cursors) != 1 {
			core.SetCursors(0, col, 0, col+1)(commandEditor)
		}
		command := string(commandEditor.Buffer.GetLine(0))
		renderer.RenderCommand(command)
		//printCommand(renderer, command)
		event := getEvent()
		for event.EventType != input.KeyPressed {
			event = getEvent()
		}
		switch event.Chr {
		case input.ENTER:
			statusText, statusOk = runCommand(command)
			renderer.ChangeStatus(statusText, statusOk)
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
		return `command "` + args[0] + `" not found`, false
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
		err := core.SaveToFile(editor, editor.Global["Filename"].(string))
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
	"!": func(args []string) (string, bool) {
		command := exec.Command("/bin/sh", "-c", strings.Join(args, " ")[1:])
		command.Stdin = core.NewEditorReader(editor, 0, 0)
		command.Stderr = io.Discard
		stdout, err := command.Output()
		if err != nil {
			return err.Error(), false
		}
		editor.MarkUndo()
		// FIXME: This is a horribly unefficient way to do this
		for editor.Buffer.GetLength() > 0 {
			editor.Buffer.RemoveLine(0)
		}
		for i, line := range strings.Split(string(stdout), "\n") {
			editor.Buffer.AddLine(i, []rune(line))
		}
		return "", false
	},
	"cursor": func(args []string) (string, bool) {
		cursor := editor.Cursors[len(editor.Cursors)-1]
		return fmt.Sprintf(
			"from row %v, col %v to row %v, col %v",
			cursor.Start.Row,
			cursor.Start.Column,
			cursor.End.Row,
			cursor.End.Column,
		), true
	},
	"treesitter": func([]string) (string, bool) {
		cursor := editor.Cursors[len(editor.Cursors)-1]
		captures := buffer.GetCaptures(cursor.Start.Row, cursor.Start.Row+1)
		output := ""
		for _, c := range captures[0] {
			output += buffer.GetCaptureName(c) + " "
		}
		return output, true
	},
	"language": func(args []string) (string, bool) {
		if len(args) != 2 {
			return "please provide exactly one language", false
		}
		lang, err := treesitter.GetLanguage(args[1])
		if err != nil {
			return err.Error(), false
		}
		buffer.SetLanguage(*lang)
		return "set language to " + args[1], true
	},
	"syntax": func([]string) (string, bool) {
		syntaxOn = !syntaxOn
		return "", true
	},
}
