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

func (b *Buffer) GetLine(index int) []rune {
	line := make([]rune, len(b.Lines[index]))
	copy(line, b.Lines[index])
	return line
}

func (b *Buffer) GetLength() int {
	return len(b.Lines)
}
