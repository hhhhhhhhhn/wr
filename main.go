package main

import (
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/hhhhhhhhhn/hexes/input"
	"github.com/hhhhhhhhhn/wr/core"
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

func main() {
	f := getFlags()
	buffer := loadBuffer(f.file)
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
	renderer = tui.NewTui()

	normalMode()
}

func loadBuffer(filename string) *core.Buffer {
	buffer := core.NewBuffer()
	contents, err := os.ReadFile(filename)
	if err == nil {
		lines := strings.Split(string(contents), "\n")
		for i, line := range lines {
			// NOTE: Equivalent to buffer.AddLine, but faster
			buffer.Current = buffer.Current.Insert(i, [][]rune{[]rune(line)})
		}
	} else {
		buffer.Current = buffer.Current.Insert(0, [][]rune{{}})
	}
	if buffer.GetLength() > 0 {
		buffer.RemoveLine(buffer.GetLength()-1)
	}
	return buffer
}
