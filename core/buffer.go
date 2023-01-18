package core

import (
	. "github.com/hhhhhhhhhn/rope"
)

type Buffer interface {
	AddLine(index int, line []rune)
	RemoveLine(index int)
	ChangeLine(index int, line []rune)
	GetLine(index int) []rune
	GetLength() int
	Backup(destination Version)
	Restore(source Version)
}

type BufferValue = *Rope[[]rune]
type Version = int

type BaseBuffer struct {
	Uri      string
	Current  BufferValue
	Versions map[Version]BufferValue
}

func (b *BaseBuffer) AddLine(index int, line []rune) {
	lineCopy := make([]rune, len(line))
	copy(lineCopy, line)
	b.Current = b.Current.Insert(index, [][]rune{line})
}

func (b *BaseBuffer) RemoveLine(index int) {
	b.Current = b.Current.Remove(index, index+1)
}

func (b *BaseBuffer) ChangeLine(index int, line []rune) {
	lineCopy := make([]rune, len(line))
	copy(lineCopy, line)
	b.Current = b.Current.Replace(index, [][]rune{line})
}

// NOTE: Be VERY CAREFUL, for performance reasons, this sends the line
// by reference, so mutating it will change the actual buffer
func (b *BaseBuffer) GetLine(index int) []rune {
	return b.Current.Slice(index, index + 1)[0]
}

func (b *BaseBuffer) GetLength() int {
	return b.Current.Length()
}

func (b *BaseBuffer) Backup(destination Version) {
	b.Versions[destination] = b.Current
}

func (b *BaseBuffer) Restore(source Version) {
	b.Current = b.Versions[source]
}

var _ Buffer = (*BaseBuffer)(nil) // Type Checking

func NewBuffer() *BaseBuffer {
	return &BaseBuffer{
		Current: NewRope([][]rune{}, DefaultSettings),
		Versions: make(map[Version]BufferValue),
	}
}
