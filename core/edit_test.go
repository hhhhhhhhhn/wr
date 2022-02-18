package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSplit(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{1,2},Location{1,3}}),
	)
	e.Do(
		PushCursor(&Range{Location{2,3},Location{2,4}}),
	)
	e.CursorDo(Split())

	expected := []string{"0000", "11", "11", "222", "2", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestSplitOOBMultiline(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{1,1},Location{2,2}}),
	)
	e.CursorDo(Split())

	expected := []string{"0000", "1", "111", "2222", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestSplitOOB(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{1,100},Location{1,200}}),
	)
	e.Do(
		PushCursor(&Range{Location{2,300},Location{2,400}}),
	)
	e.CursorDo(Split())

	expected := []string{"0000", "1111", "", "2222", "", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestSplitEOF(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{1,1}, Location{1,4}}),
	)
	e.Do(
		PushCursor(&Range{Location{2,1}, Location{2,4}}),
	)
	e.Do(
		PushCursor(&Range{Location{3,1}, Location{3,4}}),
	)
	e.CursorDo(Split())

	expected := []string{"0000", "1", "111", "2" ,"222", "3", "333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestSplitCursors(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{0,1},Location{0,4}}),
	)
	e.Do(
		PushCursor(&Range{Location{1,1},Location{1,4}}),
	)
	e.Do(
		PushCursor(&Range{Location{2,1},Location{2,4}}),
	)
	e.Do(
		PushCursor(&Range{Location{3,1},Location{3,4}}),
	)
	e.CursorDo(Split())

	expected := []string{"0", "000", "1", "111", "2", "222", "3", "333"}

	expectedCursors := []*Range{
		{Location{1,0},Location{1,3}},
		{Location{3,0},Location{3,3}},
		{Location{5,0},Location{5,3}},
		{Location{7,0},Location{7,3}},
	}

	assert.Equal(t, expected, e.Buffer.Lines)
	assert.Equal(t, expectedCursors, e.Cursors)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestSingleInsertInLine(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{1,2},Location{1,3}}),
	)
	e.Do(SingleInsertInLine("!", 1, 2))

	expected := []string{"0000", "11!11", "2222", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestInsertInLine(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{1,2},Location{1,3}}),
	)
	e.Do(
		PushCursor(&Range{Location{2,3},Location{2,4}}),
	)
	e.CursorDo(InsertInLine("!\t"))
	e.CursorDo(InsertInLine("!"))

	expected := []string{"0000", "11!\t!11", "222!\t!2", "3333"}

	expectedCursors := []*Range{
		{Location{1,8},Location{1,9}},
		{Location{2,9},Location{2,10}},
	}

	assert.Equal(t, expected, e.Buffer.Lines)
	assert.Equal(t, expectedCursors, e.Cursors)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestInsert(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{1,2},Location{1,3}}),
	)
	e.Do(
		PushCursor(&Range{Location{2,3},Location{2,4}}),
	)
	e.CursorDo(Insert("!\n!"))

	expected := []string{"0000", "11!", "!11", "222!", "!2", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)

}
func TestInsertOOB(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{1,100},Location{1,100}}),
	)
	e.Do(
		PushCursor(&Range{Location{2,100},Location{2,100}}),
	)
	e.CursorDo(Insert("!\n!"))

	expected := []string{"0000", "1111!", "!", "2222!", "!", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestSingleDelete(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.Do(
		UndoMarker(),
	)
	e.Do(SingleDelete(Range{Location{1, 1}, Location{1, 2}}))

	expected := []string{"0000", "111", "2222", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestSingleDeleteMultiline(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.Do(
		UndoMarker(),
	)
	e.Do(SingleDelete(Range{Location{1, 1}, Location{2, 2}}))

	expected := []string{"0000", "122", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestSingleDeleteJoin(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.Do(
		UndoMarker(),
	)
	e.Do(SingleDelete(Range{Location{1, 1}, Location{2, 5}}))

	expected := []string{"0000", "13333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestSingleDeleteLastLine(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.Do(
		UndoMarker(),
	)
	e.Do(SingleDelete(Range{Location{3, 4}, Location{3, 5}}))

	expected := []string{"0000", "1111", "2222", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestSingleDeleteCursors(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{1,3},Location{1,4}}),
	)
	e.Do(
		PushCursor(&Range{Location{2,3},Location{2,4}}),
	)
	e.Do(SingleDelete(Range{Location{0, 0}, Location{1, 1}}))

	expected := []string{"111", "2222", "3333"}

	expectedCursors := []*Range{
		{Location{0,2},Location{0,3}},
		{Location{1,3},Location{1,4}},
	}

	assert.Equal(t, expected, e.Buffer.Lines)
	assert.Equal(t, expectedCursors, e.Cursors)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestDelete(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{1,2},Location{1,3}}),
	)
	e.Do(
		PushCursor(&Range{Location{2,3},Location{2,4}}),
	)
	e.CursorDo(Delete())

	expected := []string{"0000", "111", "222", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestDeleteMultiline(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{1,2},Location{2,3}}),
	)
	e.CursorDo(Delete())

	expected := []string{"0000", "112", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestDeleteJoinLines(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{1,4},Location{2,5}}),
	)
	e.CursorDo(Delete())

	expected := []string{"0000", "11113333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestDeleteJoinLinesMultiline(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{0,1},Location{0,5}}),
	)
	e.Do(
		PushCursor(&Range{Location{1,1},Location{1,5}}),
	)
	e.Do(
		PushCursor(&Range{Location{2,1},Location{2,5}}),
	)
	e.Do(
		PushCursor(&Range{Location{3,1},Location{3,5}}),
	)
	e.CursorDo(Delete())

	expected := []string{"0123"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}

func TestDeleteOOB(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.Do(
		UndoMarker(),
	)
	e.Do(
		PushCursor(&Range{Location{1,100},Location{2,5}}),
	)
	e.CursorDo(Delete())

	expected := []string{"0000", "11113333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}
