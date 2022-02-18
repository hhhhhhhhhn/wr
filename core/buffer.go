package core

type Buffer struct {
	Uri   string
	Lines [][]rune
}

func (b *Buffer) AddLine(index int, line []rune) {
	lineCopy := make([]rune, len(line))
	copy(lineCopy, line)

	b.Lines = append(b.Lines, nil)
	copy(b.Lines[index+1:], b.Lines[index:])
	b.Lines[index] = lineCopy
}

func (b *Buffer) RemoveLine(index int) {
	b.Lines = append(b.Lines[:index], b.Lines[index+1:]...)
}

func (b *Buffer) ChangeLine(index int, line []rune) {
	lineCopy := make([]rune, len(line))
	copy(lineCopy, line)
	b.Lines[index] = lineCopy
}

// NOTE: Be VERY CAREFUL, for performance reasons, this sends the line
// by reference, so mutating it will change the actual buffer
func (b *Buffer) GetLine(index int) []rune {
	return b.Lines[index]
}

func (b *Buffer) GetLength() int {
	return len(b.Lines)
}
