package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/hhhhhhhhhn/hexes"
	"github.com/hhhhhhhhhn/hexes/input"
	"github.com/hhhhhhhhhn/wr/core"
)

var scroll = 0
var out = bufio.NewWriterSize(os.Stdout, 4096)
var in  = bufio.NewReader(os.Stdin)
var editor core.Editor
var renderer *hexes.Renderer
var listener  *input.Listener

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

func normalGetMovement() (movement core.Movement, ok bool) {
	event := getEvent()
	for event.EventType != input.KeyPressed {
		event = getEvent()
	}
	switch event.Chr {
	case 'l':
		return core.Chars(1), true
	case 'h':
		return core.Chars(-1), true
	case 'L':
		return core.Columns(1), true
	case 'H':
		return core.Columns(-1), true
	case 'j':
		return core.Rows(1), true
	case 'k':
		return core.Rows(-1), true
	default:
		unGetEvent()
		return nil, false
	}
}

func visualGetMovement() (movement core.Movement, ok bool) {
	event := getEvent()
	for event.EventType != input.KeyPressed {
		event = getEvent()
	}
	switch event.Chr {
	case 'L':
		return core.Chars(1), true
	case 'H':
		return core.Chars(-1), true
	case 'l':
		return core.Columns(1), true
	case 'h':
		return core.Columns(-1), true
	case 'j':
		return core.Rows(1), true
	case 'k':
		return core.Rows(-1), true
	default:
		unGetEvent()
		return nil, false
	}
}

func normalMode() {
	for {
		for len(editor.Cursors) == 0 {
			editor.SingleUndo()
		}
		PrintEditor(&editor, renderer)

		movement, ok := normalGetMovement()
		if ok {
			editor.Do(core.GoTo(movement))
			continue
		}

		event := getEvent()
		if event.EventType != input.KeyPressed {
			continue
		}

		switch(event.Chr) {
		case 'u':
			editor.Undo()
			break
		case 'U':
			editor.MarkUndo()
			break
		case 18: // <C-r>
			editor.Redo()
			break
		case 23: // <C-w>
			renderer.End()
			out.Flush()
			os.Exit(0)
			break
		case 12: // <C-l>
			renderer.Refresh()
			out.Flush()
			break
		case 'v':
			visualMode()
			break
		case 'i':
			editor.MarkUndo()
			insertMode()
			break
		case 'd':
			if movement, ok := normalGetMovement(); ok {
				editor.MarkUndo()
				editor.Do(core.SelectUntil(movement))
				editor.CursorDo(core.Delete())
			}
		}
	}
}

func visualMode() {
	for {
		for len(editor.Cursors) == 0 {
			editor.SingleUndo()
		}
		PrintEditor(&editor, renderer)

		movement, ok := normalGetMovement()
		if ok {
			editor.Do(core.ExpandSelection(movement))
			continue
		}

		event := getEvent()
		if event.EventType != input.KeyPressed {
			continue
		}

		switch(event.Chr) {
		case 'u':
			editor.Undo()
			break
		case 'U':
			editor.MarkUndo()
			break
		case 18: // <C-r>
			editor.Redo()
			break
		case 23: // <C-w>
			renderer.End()
			out.Flush()
			os.Exit(0)
			break
		case 12: // <C-l>
			renderer.Refresh()
			out.Flush()
			break
		case 'i':
			editor.MarkUndo()
			insertMode()
			break
		case input.ESCAPE:
			return
		case 'd':
			editor.MarkUndo()
			editor.CursorDo(core.Delete())
		}
	}
}

func insertMode() {
	for {
		PrintEditor(&editor, renderer)
		event := getEvent()
		if event.EventType != input.KeyPressed {
			continue
		}
		switch(event.Chr) {
		case input.ESCAPE:
			return
		case input.BACKSPACE:
			editor.CursorDo(core.MoveChars(-1))
			editor.CursorDo(core.Delete())
			break
		default:
			if unicode.IsGraphic(event.Chr) || event.Chr == '\t' || event.Chr == '\n' {
				editor.CursorDo(core.Insert([]rune{event.Chr}))
			} else {
				editor.CursorDo(core.Insert([]rune(fmt.Sprint(event.Chr))))
			}
			break
		}
	}
}

func main() {
	editor = core.Editor{Buffer: &core.Buffer{Lines: core.ToRune([]string{"aaaa", "bbbb", "cccc"})}, Config: core.EditorConfig{Tabsize: 4}}
	renderer = hexes.New(os.Stdin, out)
	listener = input.New(in)
	renderer.Start()
	editor.Do(
		core.PushCursor(&core.Range{
			Start: core.Location{Row: 0, Column: 0},
			End: core.Location{Row: 0, Column: 1}},
		),
	)

	PrintEditor(&editor, renderer)
	normalMode()
}

func PrintEditor(e *core.Editor, r *hexes.Renderer) {
	lineAmount := e.Buffer.GetLength()

	var row int
	for row = scroll; row < scroll + r.Rows && row < lineAmount; row++ {
		line := strings.ReplaceAll(string(e.Buffer.GetLine(row)), "\t", strings.Repeat(" ", e.Config.Tabsize))
		columnSpan := core.ColumnSpan(e, []rune(line))

		if columnSpan < r.Cols {
			line += strings.Repeat(" ", r.Cols - columnSpan)
		}

		col := 0
		for _, chr := range line {
			if isWithinCursor(e, row, col) {
				r.SetAttribute(hexes.REVERSE)
			} else {
				r.SetAttribute(r.DefaultAttribute)
			}
			r.SetString(row - scroll, col, string(chr))
			col += core.RuneWidth(e, chr)
		}
	}

	for ;row < scroll + r.Rows; row++ {
		r.SetString(row - scroll, 0, strings.Repeat(" ", r.Cols))
	}

	out.Flush()
}

func isWithinCursor(e *core.Editor, row, col int) bool {
	for _, cursor := range e.Cursors {
		if (
			((row == cursor.Start.Row && col >= cursor.Start.Column) || (row > cursor.Start.Row)) &&
			((row == cursor.End.Row && col < cursor.End.Column) || (row < cursor.End.Row))){
				return true
			}
	}
	return false
}
