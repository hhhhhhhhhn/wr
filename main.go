package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/hhhhhhhhhn/hexes/input"
	"github.com/hhhhhhhhhn/wr/core"
	"github.com/hhhhhhhhhn/wr/treesitter"
	"github.com/hhhhhhhhhn/wr/tui"
)

var scroll = 0
var editor *core.Editor
var renderer tui.Renderer
var listener *input.Listener

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

var printCapturesToStderr func()

func main() {
	f := getFlags()
	buffer := treesitter.NewBuffer()
	printCapturesToStderr = func() {fmt.Fprintln(os.Stderr, buffer.String(), "\n\n", buffer.GetCaptures(0, 10))} // DEBUG
	loadBuffer(f.file, buffer)
	editor = &core.Editor{
		Buffer: buffer,
		Config: core.EditorConfig{Tabsize: 4},
		Global: map[string]any{
			"Regex": regexp.MustCompile(`^\s(?P<Cursor>)\S`),
			"Filename": f.file,
		},
	}
	listener = input.New(os.Stdin)
	core.SetCursors(0, 0, 0, 1)(editor)
	renderer = treesitter.NewTui(buffer)

	normalMode()
}

func loadBuffer(filename string, buffer core.Buffer) {
	contents, err := os.ReadFile(filename)
	if err == nil {
		lines := strings.Split(string(contents), "\n")
		for i, line := range lines {
			buffer.AddLine(i, []rune(line))
		}
	} else {
		buffer.AddLine(0, []rune{})
	}
	if buffer.GetLength() > 0 {
		buffer.RemoveLine(buffer.GetLength()-1)
	}
}
