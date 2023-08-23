package treesitter

import (
	"github.com/hhhhhhhhhn/wr/advancedtui"
	"github.com/hhhhhhhhhn/hexes"
)

type SyntaxProvider struct {
	buffer             *Buffer
	captureToAttribute func(string) hexes.Attribute
}

func NewSyntaxProvider(buffer *Buffer, captureToAttribute func(string) hexes.Attribute) *SyntaxProvider {
	return &SyntaxProvider {
		buffer: buffer,
		captureToAttribute: captureToAttribute,
	}
}

func (s *SyntaxProvider) BeforeRender() {
	s.buffer.UpdateTreesitter()
}

func (s *SyntaxProvider) GetHighlights(startline int, endline int) [][]advancedtui.Highlight {
	capturesByLine := s.buffer.GetCaptures(startline, endline)
	highlights := [][]advancedtui.Highlight{}
	for _, lineCaptures := range capturesByLine {
		lineHighlights := []advancedtui.Highlight{}
		for _, capture := range lineCaptures {
			highlight := advancedtui.Highlight {
				Row: int(capture.Node.StartPoint().Row),
				Byte: int(capture.Node.StartPoint().Column),
				Attribute: s.captureToAttribute(s.buffer.query.CaptureNameForId(capture.Index)),
			}
			lineHighlights = append(lineHighlights, highlight)
		}
		highlights = append(highlights, lineHighlights)
	}
	return highlights
}

var _ advancedtui.SyntaxProvider = &SyntaxProvider{}
