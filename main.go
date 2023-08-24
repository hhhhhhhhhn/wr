package main

import (
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/hhhhhhhhhn/hexes/input"
	"github.com/hhhhhhhhhn/hexes"
	"github.com/hhhhhhhhhn/wr/core"
	"github.com/hhhhhhhhhn/wr/tui"
	"github.com/hhhhhhhhhn/wr/treesitter"
	"github.com/hhhhhhhhhn/wr/advancedtui"
)

var scroll = 0
var editor         *core.Editor
var renderer       *advancedtui.Tui
var listener       *input.Listener
var buffer         *treesitter.Buffer
var syntaxProvider advancedtui.SyntaxProvider

var _ tui.Renderer = renderer
var _ core.Buffer = buffer

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
	lang, _ := treesitter.GetLanguage("c")
	buffer = treesitter.NewBuffer(*lang)
	loadBuffer(f.file, buffer)
	editor = &core.Editor{
		Buffer: buffer,
		Config: core.EditorConfig{Tabsize: 4},
		Global: map[string]any{
			"Regex": regexp.MustCompile(`^\s(?P<Cursor>)\S`),
			"Filename": f.file,
		},
	}
	syntaxProvider = treesitter.NewSyntaxProvider(buffer, getAttribute)
	listener = input.New(os.Stdin)
	core.SetCursors(0, 0, 0, 1)(editor)
	renderer = advancedtui.NewTui()
	renderer.SetSyntaxProvider(syntaxProvider)

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

func getAttribute(name string) hexes.Attribute {
	if strings.HasPrefix(name, "type") {
		return hexes.Join(hexes.NORMAL, hexes.YELLOW)
	} else if strings.HasPrefix(name, "string") {
		return hexes.Join(hexes.NORMAL, hexes.BLUE, hexes.ITALIC)
	} else if strings.HasPrefix(name, "keyword") {
		return hexes.Join(hexes.NORMAL, hexes.GREEN)
	} else if strings.HasPrefix(name, "comment") {
		return hexes.Join(hexes.NORMAL, hexes.BLACK, hexes.BOLD, hexes.ITALIC)
	} else if strings.HasPrefix(name, "number") {
		return hexes.Join(hexes.NORMAL, hexes.CYAN)
	} else if strings.HasPrefix(name, "property") {
		return hexes.Join(hexes.NORMAL, hexes.BLUE)
	}
	return hexes.NORMAL
}

