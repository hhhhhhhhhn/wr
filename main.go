package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
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

func getMultiplier() int {
	multiplier := 0
	for {
		event := getEvent()
		for event.EventType != input.KeyPressed {
			event = getEvent()
		}
		switch event.Chr {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
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
	case 'L':
		return core.Columns(multiplier), true
	case 'H':
		return core.Columns(-multiplier), true
	case 'j', 'J':
		return core.Rows(multiplier), true
	case 'k', 'K':
		return core.Rows(-multiplier), true
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
	case 'l':
		return core.Columns(multiplier), true
	case 'h':
		return core.Columns(-multiplier), true
	case 'j', 'J':
		return core.Rows(multiplier), true
	case 'k', 'K':
		return core.Rows(-multiplier), true
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
		case '\n': // Also <C-j>
			editor.Do(core.PushCursorBelow())
			break
		case 11: // <C-k>
			editor.Do(core.RemoveCursor(editor.Cursors[len(editor.Cursors)-1]))
			break
		case input.ESCAPE:
			cursors := len(editor.Cursors)
			for i := 0; i < cursors - 1; i++ {
				editor.Do(core.RemoveCursor(editor.Cursors[0]))
			}
			break
		case 12: // <C-l>
			renderer.Refresh()
			out.Flush()
			break
		case 'v':
			visualMode()
			break
		case 'm':
			memoryProfile()
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

		movement, ok := visualGetMovement()
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
	insertion := []rune{}
	editor.MarkUndo()
	cursors := len(editor.Cursors)
	for i := 0; i < cursors - 20; i ++ {
		editor.Do(core.RemoveCursor(editor.Cursors[0]))
	}

	for {
		PrintEditor(&editor, renderer)
		event := getEvent()
		if event.EventType != input.KeyPressed {
			continue
		}
		switch(event.Chr) {
		case input.ESCAPE:
			editor.Undo()
			editor.CursorDo(core.Insert(insertion))
			return
		case input.BACKSPACE:
			if len(insertion) > 0 {
				insertion = insertion[:len(insertion)-1]
				editor.Do(core.GoTo(core.Chars(-1)))
				editor.CursorDo(core.Delete())
			}
			break
		default:
			if unicode.IsGraphic(event.Chr) || event.Chr == '\t' || event.Chr == '\n' {
				insertion = append(insertion, event.Chr)
				editor.CursorDo(core.Insert([]rune{event.Chr}))
			} else {
				insertion = append(insertion, []rune(fmt.Sprint(event.Chr))...)
				editor.CursorDo(core.Insert([]rune(fmt.Sprint(event.Chr))))
			}
			break
		}
	}
}

func memoryProfile() {
	file, _ := os.Create("memprof")
	defer file.Close()

	runtime.GC()
	pprof.WriteHeapProfile(file)
}

func main() {
	buffer := core.NewBuffer()
	buffer.Current = buffer.Current.Insert(0, [][]rune{{'a', 'b', 'c'}})
	editor = core.Editor{Buffer: buffer, Config: core.EditorConfig{Tabsize: 4}}
	renderer = hexes.New(os.Stdin, out)
	listener = input.New(in)
	renderer.Start()
	editor.Do(
		core.PushCursor(&core.Range{
			Start: core.Location{Row: 0, Column: 0},
			End: core.Location{Row: 0, Column: 1}},
		),
	)

	normalMode()
}

func PrintEditor(e *core.Editor, r *hexes.Renderer) {
	lastCursorRow := e.Cursors[len(e.Cursors) - 1].Start.Row
	if lastCursorRow < scroll {
		scroll = lastCursorRow
	}
	if lastCursorRow > scroll + r.Rows - 1 {
		scroll = lastCursorRow - r.Rows + 1
	}

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
			withinCursor, withinLast := isWithinCursor(e, row, col)
			if withinCursor {
				if withinLast {
					r.SetAttribute(hexes.MAGENTA + hexes.REVERSE)
				} else {
					r.SetAttribute(hexes.REVERSE)
				}
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

func isWithinCursor(e *core.Editor, row, col int) (isWithin bool, isLast bool) {
	for i, cursor := range e.Cursors {
		if (
			((row == cursor.Start.Row && col >= cursor.Start.Column) || (row > cursor.Start.Row)) &&
			((row == cursor.End.Row && col < cursor.End.Column) || (row < cursor.End.Row))){
				if i == len(e.Cursors) - 1 {
					return true, true
				}
				return true, false
			}
	}
	return false, false
}
