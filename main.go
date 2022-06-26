package main

import (
	"bufio"
	"os"
	"runtime"
	"runtime/pprof"

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

var attrDefault = hexes.NORMAL
var attrCursor = hexes.REVERSE
var attrActive = hexes.Join(hexes.MAGENTA, hexes.REVERSE)
var attrStatus = hexes.REVERSE

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
