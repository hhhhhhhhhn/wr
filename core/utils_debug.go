package core

import "fmt"
import "strings"

const RESET = "\033[0m"
const INVERT = "\033[7m"

func PrintEditor(e *Editor) {
	lineAmount := e.Buffer.GetLength()

	lines := []string{}

	for lineNumber := 0; lineNumber < lineAmount; lineNumber++ {
		lines = append(lines, strings.ReplaceAll(e.Buffer.GetLine(lineNumber), "\t", strings.Repeat(" ", e.Config.Tabsize)) + " ")
	}

	sortedCursors := SortCursors(e.Cursors)
	
	for i := len(sortedCursors) - 1; i >= 0; i-- {
		cursor := sortedCursors[i]
		lines[cursor.End.Row] = 
			lines[cursor.End.Row][:cursor.End.Column] +
			RESET +
			lines[cursor.End.Row][cursor.End.Column:]
		lines[cursor.Start.Row] = 
			lines[cursor.Start.Row][:cursor.Start.Column] +
			INVERT +
			lines[cursor.Start.Row][cursor.Start.Column:]
	}

	fmt.Println("-------------------------")
	for _, line := range lines {
		fmt.Println(line)
	}
	fmt.Println("-------------------------")
	fmt.Println("")
}

func isWithinACursor(e *Editor) {
}

