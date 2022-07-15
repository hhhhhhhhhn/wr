package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplit(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b}
	e.MarkUndo()

	SetCursors(1,2,1,3, 2,3,2,4)(e)
	AsEdit(Split)(e)

	expected := []string{"0000", "11", "11", "222", "2", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestSplitOOBMultiline(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b}
	e.MarkUndo()

	SetCursors(1,1,2,2)(e)

	AsEdit(Split)(e)

	expected := []string{"0000", "1", "111", "2222", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestSplitOOB(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b}
	e.MarkUndo()

	SetCursors(1,100,1,200, 2,300,2,400)(e)
	AsEdit(Split)(e)

	expected := []string{"0000", "1111", "", "2222", "", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestSplitEOF(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b}
	e.MarkUndo()

	SetCursors(1,1,1,4, 2,1,2,4, 3,1,3,4)(e)

	AsEdit(Split)(e)

	expected := []string{"0000", "1", "111", "2" ,"222", "3", "333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestSplitCursors(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SetCursors(0,1,0,4, 1,1,1,4, 2,1,2,4, 3,1,3,4)(e)
	AsEdit(Split)(e)

	expected := []string{"0", "000", "1", "111", "2", "222", "3", "333"}

	expectedCursors := []*Cursor{
		{Range: Range{Location{1,0},Location{1,3}}},
		{Range: Range{Location{3,0},Location{3,3}}},
		{Range: Range{Location{5,0},Location{5,3}}},
		{Range: Range{Location{7,0},Location{7,3}}},
	}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())
	assert.Equal(t, expectedCursors, e.Cursors)

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestSingleInsertInLine(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}

	e.MarkUndo()
	SetCursors(1,2,1,3)(e)

	SingleInsertInLine([]rune("!"), 1, 2)(e)

	expected := []string{"0000", "11!11", "2222", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestInsertInLine(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SetCursors(1,2,1,3, 2,3,2,4)(e)

	AsEdit(InsertInLine([]rune("!\t")))(e)
	AsEdit(InsertInLine([]rune("!")))(e)

	expected := []string{"0000", "11!\t!11", "222!\t!2", "3333"}

	expectedCursors := []*Cursor{
		{Range: Range{Location{1,8},Location{1,9}}},
		{Range: Range{Location{2,9},Location{2,10}}},
	}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())
	assert.Equal(t, expectedCursors, e.Cursors)

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestInsert(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SetCursors(1,2,1,3, 2,3,2,4)(e)

	AsEdit(Insert([]rune("!\n!")))(e)

	expected := []string{"0000", "11!", "!11", "222!", "!2", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestInsertOOB(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SetCursors(1,100,1,100, 2,100,2,100)(e)

	AsEdit(Insert([]rune("!\n!")))(e)

	expected := []string{"0000", "1111!", "!", "2222!", "!", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestSmartSplit(t *testing.T) {
	lines := []string{"0000", " 1111", "  2222", "   3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SetCursors(1,2,1,3, 2,3,2,4)(e)

	AsEdit(SmartSplit)(e)

	expected := []string{"0000", " 1", " 111", "  2", "  222", "   3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestSingleDelete(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SingleDelete(Range{Location{1, 1}, Location{1, 2}})(e)

	expected := []string{"0000", "111", "2222", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestSingleDeleteMultiline(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SingleDelete(Range{Location{1, 1}, Location{2, 2}})(e)

	expected := []string{"0000", "122", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestSingleDeleteJoin(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SingleDelete(Range{Location{1, 1}, Location{2, 5}})(e)

	expected := []string{"0000", "13333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestSingleDeleteLastLine(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SingleDelete(Range{Location{3, 4}, Location{3, 5}})(e)

	expected := []string{"0000", "1111", "2222", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestSingleDeleteCursors(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SetCursors(1,3,1,4, 2,3,2,4)(e)

	SingleDelete(Range{Location{0, 0}, Location{1, 1}})(e)

	expected := []string{"111", "2222", "3333"}

	expectedCursors := []*Cursor{
		{Range: Range{Location{0,2},Location{0,3}}},
		{Range: Range{Location{1,3},Location{1,4}}},
	}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())
	assert.Equal(t, expectedCursors, e.Cursors)

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestDelete(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SetCursors(1,2,1,3, 2,3,2,4)(e)

	AsEdit(Delete)(e)

	expected := []string{"0000", "111", "222", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestDeleteMultiline(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SetCursors(1,2,2,3)(e)

	AsEdit(Delete)(e)

	expected := []string{"0000", "112", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestDeleteJoinLines(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SetCursors(1,4,2,5)(e)

	AsEdit(Delete)(e)

	expected := []string{"0000", "11113333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestDeleteJoinLinesMultiline(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SetCursors(0,1,0,5, 1,1,1,5, 2,1,2,5, 3,1,3,5)(e)

	AsEdit(Delete)(e)

	expected := []string{"0123"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestDeleteOOB(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SetCursors(1,100,2,5)(e)

	AsEdit(Delete)(e)

	expected := []string{"0000", "11113333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}

func TestYankAndPaste(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SetCursors(1,1,2,2)(e)

	AsEdit(Yank(0))(e)
	GoTo(StartOfLine)(e)
	AsEdit(Paste(0))(e)

	expected := []string{"0000", "111", "221111", "2222", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}
func TestYankEOL(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune(lines))
	e := &Editor{Buffer: b, Config: EditorConfig{Tabsize: 4}}
	e.MarkUndo()

	SetCursors(1,0, 1,4)(e)

	AsEdit(Yank(0))(e)
	GoTo(StartOfLine)(e)
	AsEdit(Paste(0))(e)

	expected := []string{"0000", "11111111", "2222", "3333"}

	assert.Equal(t, ToRune(expected), e.Buffer.Current.Value())

	e.Undo()
	assert.Equal(t, ToRune(linesCopy), e.Buffer.Current.Value())
}
