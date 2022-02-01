package core

type Buffer struct {
	Uri   string
	Lines []string
}

func (b *Buffer) AddLine(index int, line string) {
	b.Lines = append(b.Lines, "")
	copy(b.Lines[index+1:], b.Lines[index:])
	b.Lines[index] = line
}

func (b *Buffer) RemoveLine(index int) {
	b.Lines = append(b.Lines[:index], b.Lines[index+1:]...)
}

func (b *Buffer) ChangeLine(index int, line string) {
	b.Lines[index] = line
}

func (b *Buffer) GetLine(index int) string {
	return b.Lines[index]
}

func (b *Buffer) GetLength() int {
	return len(b.Lines)
}
