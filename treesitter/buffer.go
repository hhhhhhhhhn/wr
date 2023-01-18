package treesitter

import (
	"github.com/hhhhhhhhhn/wr/core"
	"github.com/hhhhhhhhhn/rope"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

type Buffer struct {
	base              core.Buffer
	lineBytes         *rope.Rope[int]
	lineBytesVersions map[core.Version]*rope.Rope[int]

	query             *sitter.Query
	queryCursor       *sitter.QueryCursor
	parser            *sitter.Parser
	tree              *sitter.Tree
	input             sitter.Input
}

func (b *Buffer) AddLine(index int, line []rune) {
	lineBytes := len(string(line) + "\n")
	lineStartByte := uint32(b.GetLineByteStart(index))

	b.base.AddLine(index, line)
	b.lineBytes = b.lineBytes.Insert(index, []int{lineBytes})

	b.tree.Edit(sitter.EditInput{
		StartIndex: lineStartByte,
		OldEndIndex: lineStartByte,
		NewEndIndex: lineStartByte + uint32(lineBytes),
		StartPoint: sitter.Point{
			Row: uint32(index),
			Column: 0,
		},
		OldEndPoint: sitter.Point{
			Row: uint32(index),
			Column: 0,
		},
		NewEndPoint: sitter.Point{
			Row: uint32(index) + 1,
			Column: 0,
		},
	})
}

func (b *Buffer) RemoveLine(index int) {
	oldLineBytes     := uint32(b.GetLineBytes(index))
	oldLineByteStart := uint32(b.GetLineByteStart(index))

	b.base.RemoveLine(index)
	b.lineBytes = b.lineBytes.Remove(index, index+1)

	b.tree.Edit(sitter.EditInput{
		StartIndex: oldLineByteStart,
		OldEndIndex: oldLineByteStart + oldLineBytes,
		NewEndIndex: oldLineByteStart,
		StartPoint: sitter.Point{
			Row: uint32(index),
			Column: 0,
		},
		OldEndPoint: sitter.Point{
			Row: uint32(index)+1,
			Column: 0,
		},
		NewEndPoint: sitter.Point{
			Row: uint32(index),
			Column: 0,
		},
	})
}

func (b *Buffer) ChangeLine(index int, line []rune) {
	lineByteStart := uint32(b.GetLineByteStart(index))
	oldLineBytes  := uint32(b.GetLineBytes(index))
	newLineBytes  := len(string(line) + "\n")

	b.base.ChangeLine(index, line)
	b.lineBytes = b.lineBytes.Replace(index, []int{newLineBytes})

	b.tree.Edit(sitter.EditInput{
		StartIndex: lineByteStart,
		OldEndIndex: lineByteStart + oldLineBytes,
		NewEndIndex: lineByteStart + uint32(newLineBytes),
		StartPoint: sitter.Point{
			Row: uint32(index),
			Column: 0,
		},
		OldEndPoint: sitter.Point{
			Row: uint32(index),
			Column: oldLineBytes,
		},
		NewEndPoint: sitter.Point{
			Row: uint32(index),
			Column: uint32(newLineBytes),
		},
	})
}

func (b *Buffer) GetLine(index int) []rune {
	return b.base.GetLine(index)
}

func (b *Buffer) GetLineBytes(index int) int {
	return b.lineBytes.Slice(index, index+1)[0]
}

// TODO: Optimize
func (b *Buffer) GetLineByteStart(index int) int {
	lines := b.lineBytes.Slice(0, index)
	accum := 0
	for _, line := range lines {
		accum += line
	}
	return accum
}

func (b *Buffer) GetLength() int {
	return b.base.GetLength()
}

func (b *Buffer) Backup(destination core.Version) {
	b.base.Backup(destination)
	b.lineBytesVersions[destination] = b.lineBytes
}

func (b *Buffer) Restore(source core.Version) {
	b.base.Restore(source)
	b.lineBytes = b.lineBytesVersions[source]

	b.tree = b.parser.ParseInput(nil, b.input)
}

func (b *Buffer) UpdateTreesitter() {
	b.tree = b.parser.ParseInput(b.tree, b.input)
}

func (b *Buffer) GetCaptures(startRow, endRow int) [][]sitter.QueryCapture {
	b.queryCursor.SetPointRange(
		sitter.Point{
			Row: uint32(startRow),
			Column: 0,
		},
		sitter.Point{
			Row: uint32(endRow),
			Column: 0,
		},
	)

	b.queryCursor.Exec(b.query, b.tree.RootNode())

	var captures []sitter.QueryCapture

	for true {
		match, ok := b.queryCursor.NextMatch()
		if !ok {
			break
		}
		for _, capture := range match.Captures {
			if len(captures) > 0 {
				lastCapture := captures[len(captures)-1]
				if intersects(capture, lastCapture) &&
					len(b.query.CaptureNameForId(capture.Index)) >= len(b.query.CaptureNameForId(lastCapture.Index)) {
						captures[len(captures)-1] = capture
						continue
				}
			}

			captures = append(captures, capture)
		}
	}

	var capturesByLine [][]sitter.QueryCapture

	for _, capture := range captures {
		if len(capturesByLine) == 0 {
			capturesByLine = append(capturesByLine, []sitter.QueryCapture{capture})
		} else if capturesByLine[len(capturesByLine) - 1][0].Node.StartPoint().Row < capture.Node.StartPoint().Row {
			lastLine := capturesByLine[len(capturesByLine)-1]
			lastCapture := lastLine[len(lastLine)-1]
			capturesByLine = append(capturesByLine, []sitter.QueryCapture{lastCapture, capture})
		} else {
			capturesByLine[len(capturesByLine) - 1] = append(capturesByLine[len(capturesByLine) - 1], capture)
		}
	}

	for len(capturesByLine) < endRow - startRow {
		capturesByLine = append(capturesByLine, capturesByLine[len(capturesByLine)-1])
	}

	return capturesByLine
}

func NewBuffer() *Buffer {
	query, _ := sitter.NewQuery([]byte(query), javascript.GetLanguage())

	buffer := &Buffer{}
	buffer.base              = core.NewBuffer()
	buffer.lineBytes         = rope.NewRope([]int{}, rope.DefaultSettings)
	buffer.lineBytesVersions = make(map[core.Version]*rope.Rope[int])
	buffer.query             = query
	buffer.queryCursor       = sitter.NewQueryCursor()
	buffer.parser            = sitter.NewParser()
	buffer.parser.SetLanguage(javascript.GetLanguage())
	buffer.tree              = buffer.parser.Parse(nil, []byte("\n"))
	buffer.input             = sitter.Input {
		Encoding: sitter.InputEncodingUTF8,
		Read: func(_offset uint32, position sitter.Point) []byte {
			if int(position.Row) >= buffer.GetLength() {
				return nil
			}
			line := buffer.GetLine(int(position.Row))
			return []byte((string(line) + "\n"))[position.Column:]
		},
	}

	return buffer
}

var _ core.Buffer = (*Buffer)(nil)