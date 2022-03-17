package core

import (
	. "github.com/hhhhhhhhhn/rope"
)

type BufferValue = *Rope[[]rune]
type Version = int

type Buffer struct {
	Uri      string
	Current  BufferValue
	Versions map[Version]BufferValue
}

func (b *Buffer) AddLine(index int, line []rune) {
	lineCopy := make([]rune, len(line))
	copy(lineCopy, line)
	b.Current = b.Current.Insert(index, [][]rune{line})
}

func (b *Buffer) RemoveLine(index int) {
	b.Current = b.Current.Remove(index, index+1)
}

func (b *Buffer) ChangeLine(index int, line []rune) {
	lineCopy := make([]rune, len(line))
	copy(lineCopy, line)
	b.Current = b.Current.Replace(index, [][]rune{line})
}

// NOTE: Be VERY CAREFUL, for performance reasons, this sends the line
// by reference, so mutating it will change the actual buffer
func (b *Buffer) GetLine(index int) []rune {
	return b.Current.Slice(index, index + 1)[0]
}

func (b *Buffer) GetLength() int {
	return b.Current.Length()
}

func NewBuffer() *Buffer {
	return &Buffer{
		Current: NewRope([][]rune{}, DefaultSettings),
		Versions: make(map[Version]BufferValue),
	}
}
