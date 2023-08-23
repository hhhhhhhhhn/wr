package advancedtui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/hhhhhhhhhn/hexes"
	"github.com/hhhhhhhhhn/wr/core"
	"github.com/hhhhhhhhhn/wr/tui"
)

type Renderer = tui.Renderer

type Tui struct {
	out          *bufio.Writer
	renderer     *hexes.Renderer
	scroll       int
	statusText   string
	statusOk     bool
	getAttribute func(string) hexes.Attribute
	provider     SyntaxProvider
}

func NewTui() *Tui {
	out := bufio.NewWriterSize(os.Stdout, 4096)
	in  := os.Stdin
	renderer := hexes.New(in, out)
	renderer.Start()

	return &Tui {
		renderer: renderer,
		out: out,
		scroll: 0,
		statusText: "",
		statusOk: true,
		provider: &NoHighlight{},
	}
}

func (t *Tui) SetSyntaxProvider(provider SyntaxProvider) {
	t.provider = provider
}

func (t *Tui) fillBlank() {
	t.renderer.SetAttribute(attrDefault)
	for i := 0; i < t.renderer.Rows; i++ {
		t.renderer.SetString(i, 0, strings.Repeat(" ", t.renderer.Cols))
	}
}

func (t *Tui) RenderEditor(e *core.Editor) {
	t.fillBlank()

	renderRows := t.renderer.Rows - 1 // Extra row for commands
	t.scroll = handleScroll(e, renderRows, t.scroll)

	lineAmount := e.Buffer.GetLength()
	t.provider.BeforeRender()
	highlights := t.provider.GetHighlights(t.scroll, t.scroll + renderRows)
	for row := t.scroll; row < t.scroll + renderRows && row < lineAmount; row++ {
		printLine(e, t, highlights[row - t.scroll], row, t.scroll)
	}

	printStatusBar(e, t.renderer, t.statusText, t.statusOk)

	t.out.Flush()
}

func (t *Tui) End() {
	t.renderer.End()
	t.out.Flush()
}

var attrDefault   = hexes.NORMAL
var attrStatusErr = hexes.Join(hexes.NORMAL, hexes.BOLD, hexes.BG_RED, hexes.REVERSE)
var attrCursor    = hexes.Join(hexes.NORMAL, hexes.REVERSE)
var attrActive    = hexes.Join(hexes.NORMAL, hexes.MAGENTA, hexes.REVERSE)
var attrStatus    = hexes.Join(hexes.NORMAL, hexes.REVERSE)

func handleScroll(e*core.Editor, renderRows int, currentScroll int) (newScroll int) {
	var lastCursorRow int
	if len(e.Cursors) > 0 {
		lastCursorRow = e.Cursors[len(e.Cursors) - 1].Start.Row
	} else {
		lastCursorRow = 0
	}
	if lastCursorRow < currentScroll {
		currentScroll = lastCursorRow
	}
	if lastCursorRow > currentScroll + renderRows - 1 {
		currentScroll = lastCursorRow - renderRows + 1
	}
	return currentScroll
}

func padWithSpaces(str string, cols int, desiredCols int) string { if cols < desiredCols {
		str += strings.Repeat(" ", desiredCols - cols)
	}
	return str
}

func getLineAsString(e *core.Editor, row int) string {
	return strings.ReplaceAll(string(e.Buffer.GetLine(row)), "\t", strings.Repeat(" ", e.Config.Tabsize))
}

func printLine(e *core.Editor, tui *Tui, highlights []Highlight, row, scroll int) {
	line := e.Buffer.GetLine(row)
	originalLineCols  := core.ColumnSpan(e, line)
	if originalLineCols < tui.renderer.Cols {
		line = append(line, []rune(strings.Repeat(" ", tui.renderer.Cols - originalLineCols))...)
	}

	col := 0
	byt := 0
	for _, chr := range line {
		// Advances the captures
		for len(highlights) > 1 && byt >= highlights[1].Byte {
			highlights = highlights[1:]
		}

		withinCursor, withinLast, cursor := isWithinCursor(e, row, col)
		if withinCursor && (col <= originalLineCols || (cursor.Start.Row == row && cursor.Start.Column > originalLineCols)) {
			if withinLast {
				tui.renderer.SetAttribute(attrActive)
			} else {
				tui.renderer.SetAttribute(attrCursor)
			}
		} else if len(highlights) > 0 {
			tui.renderer.SetAttribute(highlights[0].Attribute)
		}
		if chr == '\t' {
			tui.renderer.SetString(row - scroll, col, strings.Repeat(" ", e.Config.Tabsize))
		} else {
			tui.renderer.SetString(row - scroll, col, string(chr))
		}
		col += core.RuneWidth(e, chr)
		byt += utf8.RuneLen(chr)
	}

	tui.renderer.SetAttribute(attrDefault)
}

func isWithinCursor(e *core.Editor, row, col int) (isWithin bool, isLast bool, cursor *core.Cursor) {
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
					return true, true, cursor
				}
				return true, false, cursor
			}
	}
	return false, false, nil
}

func (t *Tui) ChangeStatus(text string, ok bool) {
	t.statusText = text
	t.statusOk = ok
}

func printStatusBar(e *core.Editor, r *hexes.Renderer, statusText string, statusOk bool) {
	row := r.Rows - 1
	var position string
	if len(e.Cursors) > 0 {
		position = fmt.Sprintf("line %v, col %v ",
			e.Cursors[len(e.Cursors)-1].Start.Row,
			e.Cursors[len(e.Cursors)-1].Start.Column,
		)
	}

	statusText = " " + statusText
	statusText = padWithSpaces(statusText, len(statusText), r.Cols)
	if statusOk {
		r.SetAttribute(attrStatus)
	} else {
		r.SetAttribute(attrStatusErr)
	}
	r.SetString(row, 0, statusText)
	r.SetString(row, r.Cols-len(position), position)
}

func (t *Tui) RenderCommand(command string, cursorPos int) {
	row := t.renderer.Rows - 1
	formatted := padWithSpaces(":" + command, len(command), t.renderer.Cols)
	cursorPos = cursorPos+1

	t.renderer.SetAttribute(attrStatus)
	t.renderer.SetString(row, 0, formatted[:cursorPos])
	t.renderer.SetAttribute(attrActive)
	t.renderer.SetString(row, cursorPos, formatted[cursorPos:cursorPos+1])
	t.renderer.SetAttribute(attrStatus)
	t.renderer.SetString(row, cursorPos+1, formatted[cursorPos+1:])

	t.out.Flush()
}
