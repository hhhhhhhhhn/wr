package main

import (
	"unicode"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hhhhhhhhhn/hexes/input"
	"github.com/hhhhhhhhhn/wr/core"
)

const eventsLength = 10
var events [eventsLength]*input.Event
var latestEvent = -1
var eventIndex = -1

func getEvent() *input.Event {
	eventIndex++
	for eventIndex > latestEvent {
		latestEvent++
		events[latestEvent % eventsLength] = listener.GetEvent()
	}
	return events[eventIndex % eventsLength]
}

func unGetEvent() {
	eventIndex--
}

func getMultiplier() int {
	multiplier := 0
	for {
		event := getEvent()
		for event.EventType != input.KeyPressed {
			event = getEvent()
		}
		switch event.Chr {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if event.Chr == '0' && multiplier == 0 {
				unGetEvent()
				return 1
			}
			multiplier = 10 * multiplier + int(event.Chr - '0')
			break
		default:
			unGetEvent()
			if multiplier == 0 {
				return 1
			}
			return multiplier
		}
	}
}

func getRegister() int {
	event := getEvent()
	for event.EventType != input.KeyPressed {
		event = getEvent()
	}
	if event.Chr - 'a' < 0 || event.Chr - 'a' > 30 {
		return 0
	}
	return int(event.Chr - 'a')
}

func normalGetMovement() (movement core.Movement, ok bool) {
	multiplier := getMultiplier()
	event := getEvent()
	for event.EventType != input.KeyPressed {
		event = getEvent()
	}
	switch event.Chr {
	case 'l':
		return core.Chars(multiplier), true
	case 'h':
		return core.Chars(-multiplier), true
	case 'w':
		return core.Words(multiplier), true
	case 'b':
		return core.Words(-multiplier), true
	case 'L':
		return core.Columns(multiplier), true
	case 'H':
		return core.Columns(-multiplier), true
	case 'j', 'J':
		return core.Rows(multiplier), true
	case 'k', 'K':
		return core.Rows(-multiplier), true
	case 'n':
		return core.Regex(editor.Global["Regex"].(*regexp.Regexp), multiplier), true
	case 'N':
		return core.Regex(editor.Global["Regex"].(*regexp.Regexp), -multiplier), true
	case '0':
		return core.StartOfLine, true
	case '$':
		return core.EndOfLine, true
	default:
		unGetEvent()
		return nil, false
	}
}

func visualGetMovement() (movement core.Movement, ok bool) {
	multiplier := getMultiplier()
	event := getEvent()
	for event.EventType != input.KeyPressed {
		event = getEvent()
	}
	switch event.Chr {
	case 'L':
		return core.Chars(multiplier), true
	case 'H':
		return core.Chars(-multiplier), true
	case 'w':
		return core.Words(multiplier), true
	case 'b':
		return core.Words(-multiplier), true
	case 'l':
		return core.Columns(multiplier), true
	case 'h':
		return core.Columns(-multiplier), true
	case 'j', 'J':
		return core.Rows(multiplier), true
	case 'k', 'K':
		return core.Rows(-multiplier), true
	case 'n':
		return core.Regex(editor.Global["Regex"].(*regexp.Regexp), multiplier), true
	case 'N':
		return core.Regex(editor.Global["Regex"].(*regexp.Regexp), -multiplier), true
	case '0':
		return core.StartOfLine, true
	case '$':
		return core.EndOfLine, true
	default:
		unGetEvent()
		return nil, false
	}
}

func baseActions(char rune) (ok bool) {
	switch(char) {
	case 'u':
		editor.Undo()
		return true
	case 'U':
		editor.MarkUndo()
		return true
	case 18: // <C-r>
		editor.Redo()
		return true
	case 23: // <C-w>
		err := core.SaveToFile(editor, editor.Global["Filename"].(string))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		quit()
		return true
	case 12: // <C-l>
		// TODO: Re-do with message
		//renderer.Refresh()
		//out.Flush()
		return true
	case 'v':
		visualMode()
		return true
	case 22: // <C-v>
		newCursorMode()
		return true
	case 'M':
		memoryProfile()
		return true
	case 'C':
		toggleCpuProf()
		return true
	case 'i':
		editor.MarkUndo()
		insertMode()
		return true
	case 'I':
		editor.MarkUndo()
		core.GoTo(core.StartOfLine)(editor)
		insertMode()
		return true
	case 'a':
		editor.MarkUndo()
		core.GoTo(core.Chars(1))(editor)
		insertMode()
		return true
	case 'A':
		editor.MarkUndo()
		core.GoTo(core.EndOfLine)(editor)
		insertMode()
		return true
	case 'o':
		editor.MarkUndo()
		core.GoTo(core.EndOfLine)(editor)
		core.AsEdit(core.Insert([]rune{'\n'}))(editor)
		insertMode()
		return true
	case 'O':
		editor.MarkUndo()
		core.GoTo(core.StartOfLine)(editor)
		core.AsEdit(core.Insert([]rune{'\n'}))(editor)
		core.GoTo(core.Rows(-1))(editor)
		insertMode()
		return true
	case 'd':
		if movement, ok := normalGetMovement(); ok {
			editor.MarkUndo()
			core.SelectUntil(movement)(editor)
			core.AsEdit(core.Delete)(editor)
		}
	case 'c':
		if movement, ok := normalGetMovement(); ok {
			editor.MarkUndo()
			core.SelectUntil(movement)(editor)
			core.AsEdit(core.Delete)(editor)
			insertMode()
		}
		break
	case 's':
		editor.MarkUndo()
		core.AsEdit(core.Delete)(editor)
		insertMode()
		break
	case 'x':
		editor.MarkUndo()
		core.AsEdit(core.Delete)(editor)
		break
	case 'y':
		core.AsEdit(core.Yank(getRegister()))(editor)
		break
	case 'p':
		core.AsEdit(core.Paste(getRegister()))(editor)
		break
	case ':':
		commandMode("")
	case '/':
		commandMode("/")
	}
	return false
}

func normalMode() {
	pushMode("normal")
	defer popMode()
	lastCursor := defaultCursor
	for {
		if len(editor.Cursors) == 0 {
			cursorCopy := lastCursor
			core.PushCursor(&cursorCopy)(editor)
		} else if !core.IsOOB(editor, editor.Cursors[len(editor.Cursors)-1]) {
			lastCursor = *editor.Cursors[len(editor.Cursors)-1]
		}
		renderer.RenderEditor(editor)

		movement, ok := normalGetMovement()
		if ok {
			core.GoTo(movement)(editor)
			continue
		}

		event := getEvent()
		if event.EventType != input.KeyPressed {
			continue
		}

		switch event.Chr {
		case input.ESCAPE:
			core.OnlyMainCursor(editor)
			core.GoTo(core.Unselect)(editor)
			break
		case 'g':
			if getEvent().Chr == 'g' {
				core.OnlyMainCursor(editor)
				core.GoTo(core.Position(0, 0, 0, 1))(editor)
			}
			break
		case 'G':
			length := editor.Buffer.GetLength()
			if length == 0 {
				break
			}
			core.OnlyMainCursor(editor)
			core.GoTo(core.Position(length-1, 0, length-1, 1))(editor)
		default:
			baseActions(event.Chr)
			break
		}
	}
}

func visualMode() {
	pushMode("visual")
	defer popMode()
	for {
		for len(editor.Cursors) == 0 {
			core.SetCursors(0, 0, 0, 1)(editor)
		}
		renderer.RenderEditor(editor)

		movement, ok := visualGetMovement()
		if ok {
			core.ExpandSelection(movement)(editor)
			continue
		}

		event := getEvent()
		if event.EventType != input.KeyPressed {
			continue
		}

		switch(event.Chr) {
		case input.ESCAPE:
			core.GoTo(core.Unselect)(editor)
			return
		case 'd':
			editor.MarkUndo()
			core.AsEdit(core.Delete)(editor)
			break
		default:
			baseActions(event.Chr)
			break
		}
	}
}

func newCursorMode() {
	pushMode("new cursor")
	defer popMode()
	for {
		for len(editor.Cursors) == 0 {
			core.SetCursors(0, 0, 0, 1)(editor)
		}
		renderer.RenderEditor(editor)

		movement, ok := visualGetMovement()
		if ok {
			core.PushCursorFromLast(movement)(editor)
			continue
		}

		event := getEvent()
		if event.EventType != input.KeyPressed {
			continue
		}

		switch(event.Chr) {
		case input.ESCAPE:
			return
		default:
			baseActions(event.Chr)
			break
		}
	}
}

// FIXME: Doesn't always match entered text
func insertMode() {
	pushMode("insert")
	defer popMode()

	editor.MarkUndo()
	for len(editor.Cursors) > 50 {
		core.RemoveCursor(editor.Cursors[0])(editor)
	}
	edits := []core.Edit{}
	do := func(e core.Edit) {edits = append(edits, e); e(editor)}

	for {
		renderer.RenderEditor(editor)
		event := getEvent()
		if event.EventType != input.KeyPressed {
			continue
		}
		switch(event.Chr) {
		case input.ESCAPE:
			editor.Undo()
			for _, edit := range edits {
				edit(editor)
			}
			return
		case input.BACKSPACE:
			do(core.GoTo(core.Chars(-1)))
			do(core.AsEdit(core.Delete))
			break
		case '\n':
			do(core.AsEdit(core.SmartSplit))
		default:
			if unicode.IsGraphic(event.Chr) || event.Chr == '\t' {
				do(core.AsEdit(core.Insert([]rune{event.Chr})))
			} else {
				do(core.AsEdit(core.Insert([]rune(fmt.Sprint(event.Chr)))))
			}
			break
		}
		if len(editor.Cursors) == 0 {
			do(core.SetCursors(0, 0, 0, 1))
		}
	}
}

var modes = []string{}
var statusText string
var statusOk bool = true

func pushMode(mode string) {
	modes = append(modes, mode)
	updateStatusText()
}

func popMode() {
	modes = modes[:len(modes)-1]
	updateStatusText()
}

func updateStatusText() {
	statusText = strings.Join(modes, " > ")
	statusOk = true
	renderer.ChangeStatus(statusText, statusOk)
}

var defaultCursor = core.Cursor{
	Range: core.Range{
		Start: core.Location{Row: 0, Column: 0},
		End: core.Location{Row: 0, Column: 1},
	},
}

func quit() {
	renderer.End()
	os.Exit(0)
}
