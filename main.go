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
var editor *core.Editor
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
	}
	return false
}

func normalMode() {
	for {
		for len(editor.Cursors) == 0 {
			core.SetCursors(0, 0, 0, 1)(editor)
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

func memoryProfile() {
	file, _ := os.Create("memprof")
	defer file.Close()

	runtime.GC()
	pprof.WriteHeapProfile(file)
}

var cpuProfFile *os.File
func toggleCpuProf() {
	if cpuProfFile == nil {
		cpuProfFile, _ = os.Create("cpuprof")
		pprof.StartCPUProfile(cpuProfFile)
	} else {
		pprof.StopCPUProfile()
		cpuProfFile.Close()
		cpuProfFile = nil
	}
}

func main() {
	buffer := core.NewBuffer()
	buffer.Current = buffer.Current.Insert(0, [][]rune{{'a', 'b', 'c'}})
	editor = &core.Editor{Buffer: buffer, Config: core.EditorConfig{Tabsize: 4}}
	renderer = hexes.New(os.Stdin, out)
	listener = input.New(in)
	renderer.Start()
	core.SetCursors(0, 0, 0, 1)(editor)

	normalMode()
}

func PrintEditor(e *core.Editor, r *hexes.Renderer) {
	var lastCursorRow int
	if len(editor.Cursors) > 0 {
		lastCursorRow = e.Cursors[len(e.Cursors) - 1].Start.Row
	} else {
		lastCursorRow = 0
	}
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
	var cursors []*core.Cursor
	if len(e.Cursors) > 25 {
		cursors = e.Cursors[len(e.Cursors)-25:]
	} else {
		cursors = e.Cursors
	}
	for i, cursor := range cursors {
		if (
			((row == cursor.Start.Row && col >= cursor.Start.Column) || (row > cursor.Start.Row)) &&
			((row == cursor.End.Row && col < cursor.End.Column) || (row < cursor.End.Row))){
				if i == len(cursors) - 1 {
					return true, true
				}
				return true, false
			}
	}
	return false, false
}
