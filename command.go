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
	"github.com/hhhhhhhhhn/wr/advancedtui"
)

func commandMode(command string) {
	cursor := 0

	for {
		renderer.RenderCommand(command, cursor)
		event := getEvent()
		for event.EventType != input.KeyPressed {
			event = getEvent()
		}
		switch event.Chr {
		case input.KEY_UP:
			break
		case input.KEY_DOWN:
			break
		case input.KEY_LEFT:
			 if cursor > 0 {
				 cursor--
			 }
			break
		case input.KEY_RIGHT:
			 if cursor < len(command) {
				 cursor++
			 }
			break
		case input.ENTER:
			statusText, statusOk = runCommand(command)
			renderer.ChangeStatus(statusText, statusOk)
			return
		case input.ESCAPE:
			return
		case input.BACKSPACE:
			if cursor > 0 {
				command = command[:cursor - 1] + command[cursor:]
				cursor--
			}
			break
		default:
			command = command[:cursor] + string(event.Chr) + command[cursor:]
			cursor++
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
	"syntax": func(args []string) (string, bool) {
		if len(args) != 2 {
			return "please provide exactly one provider (none or treesitter)", false
		} else if args[1] == "none" {
			renderer.SetSyntaxProvider(&advancedtui.NoHighlight{})
			return "", true
		} else if args[1] == "treesitter" {
			renderer.SetSyntaxProvider(syntaxProvider)
			return "", true
		} else {
			return "unknown syntax provider: " + args[1], false
		}
	},
}
