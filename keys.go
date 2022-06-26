package main

import (
	"unicode"
	"fmt"
	"os"

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
		err := core.SaveToFile(editor, "file.txt")
		renderer.End()
		out.Flush()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(0)
		return true
	case 12: // <C-l>
		renderer.Refresh()
		out.Flush()
		return true
	case 'v':
		visualMode()
		return true
	case 22: // <C-v>
		newCursorMode()
		return true
	case 'm':
		memoryProfile()
		return true
	case 'c':
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
	case ':':
		commandMode()
	}
	return false
}

var defaultCursor = core.Cursor{
	Range: core.Range{
		Start: core.Location{Row: 0, Column: 0},
		End: core.Location{Row: 0, Column: 1},
	},
	Registers: []string{},
}


func normalMode() {
	lastCursor := defaultCursor
	for {
		if len(editor.Cursors) == 0 {
			cursorCopy := lastCursor
			core.PushCursor(&cursorCopy)(editor)
		} else if !core.IsOOB(editor, editor.Cursors[len(editor.Cursors)-1]) {
			lastCursor = *editor.Cursors[len(editor.Cursors)-1]
		}
		PrintEditor(editor, renderer)

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
			cursors := len(editor.Cursors)
			for i := 0; i < cursors - 1; i++ {
				core.RemoveCursor(editor.Cursors[0])(editor)
			}
			break
		default:
			baseActions(event.Chr)
		}
	}
}

func visualMode() {
	for {
		for len(editor.Cursors) == 0 {
			core.SetCursors(0, 0, 0, 1)(editor)
		}
		PrintEditor(editor, renderer)

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
	for {
		for len(editor.Cursors) == 0 {
			core.SetCursors(0, 0, 0, 1)(editor)
		}
		PrintEditor(editor, renderer)

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
	insertion := []rune{}
	deletion := 0
	editor.MarkUndo()
	cursors := len(editor.Cursors)
	for i := 0; i < cursors - 20; i ++ {
		core.RemoveCursor(editor.Cursors[0])(editor)
	}

	for {
		PrintEditor(editor, renderer)
		event := getEvent()
		if event.EventType != input.KeyPressed {
			continue
		}
		switch(event.Chr) {
		case input.ESCAPE:
			editor.Undo()
			core.GoTo(core.Chars(-deletion))(editor)
			core.SelectUntil(core.Chars(deletion))(editor)
			core.AsEdit(core.Delete)(editor)
			core.AsEdit(core.Insert(insertion))(editor)
			return
		case input.BACKSPACE:
			if len(insertion) == 0 {
				deletion++
			} else {
				insertion = insertion[:len(insertion)-1]
			}
			core.GoTo(core.Chars(-1))(editor)
			core.AsEdit(core.Delete)(editor)
			break
		default:
			if unicode.IsGraphic(event.Chr) || event.Chr == '\t' || event.Chr == '\n' {
				insertion = append(insertion, event.Chr)
				core.AsEdit(core.Insert([]rune{event.Chr}))(editor)
			} else {
				insertion = append(insertion, []rune(fmt.Sprint(event.Chr))...)
				core.AsEdit(core.Insert([]rune(fmt.Sprint(event.Chr))))(editor)
			}
			break
		}
		if len(editor.Cursors) == 0 {
			core.SetCursors(0, 0, 0, 1)(editor)
		}
	}
}
