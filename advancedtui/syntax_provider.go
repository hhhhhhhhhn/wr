package advancedtui

import (
	"github.com/hhhhhhhhhn/hexes"
)

type Highlight struct {
	Row       int
	Byte      int // Within row
	Attribute hexes.Attribute
}

type SyntaxProvider interface {
	BeforeRender()
	GetHighlights(startline int, endline int) [][]Highlight
}

type NoHighlight struct {}
func (n *NoHighlight) BeforeRender() {}
func (n *NoHighlight) GetHighlights(startline int, endline int) [][]Highlight {
	highlights := [][]Highlight{}
	for i := startline; i < endline; i++ {
		highlights = append(highlights, []Highlight{{Row: i, Byte: 0, Attribute: hexes.NORMAL}})
	}
	return highlights
}
var _ SyntaxProvider = &NoHighlight{}
