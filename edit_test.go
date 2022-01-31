package main

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
	PrintEditor(&e)
	e.CursorDo(Split())
	PrintEditor(&e)

	expected := []string{"0000", "11", "11", "222", "2", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	PrintEditor(&e)
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
	PrintEditor(&e)
	e.Do(SingleInsertInLine("!", 1, 2))
	PrintEditor(&e)

	expected := []string{"0000", "11!11", "2222", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	PrintEditor(&e)
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
	PrintEditor(&e)
	e.CursorDo(InsertInLine("!\t"))
	PrintEditor(&e)
	e.CursorDo(InsertInLine("!"))
	PrintEditor(&e)

	expected := []string{"0000", "11!\t!11", "222!\t!2", "3333"}

	expectedCursors := []*Range{
		{Location{1,8},Location{1,9}},
		{Location{2,9},Location{2,10}},
	}

	assert.Equal(t, expected, e.Buffer.Lines)
	assert.Equal(t, expectedCursors, e.Cursors)

	e.Undo()
	PrintEditor(&e)
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

func TestSingleDelete(t *testing.T) {
	lines := []string{"0000", "1111", "2222", "3333"}
	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	b := Buffer{Lines: lines}
	e := Editor{Buffer: &b, Config: EditorConfig{Tabsize: 4}}

	e.Do(
		UndoMarker(),
	)
	PrintEditor(&e)
	e.Do(SingleDelete(Range{Location{1, 1}, Location{1, 2}}))
	PrintEditor(&e)

	expected := []string{"0000", "111", "2222", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	PrintEditor(&e)
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
	PrintEditor(&e)
	e.Do(SingleDelete(Range{Location{1, 1}, Location{2, 2}}))
	PrintEditor(&e)

	expected := []string{"0000", "122", "3333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	PrintEditor(&e)
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
	PrintEditor(&e)
	e.Do(SingleDelete(Range{Location{1, 1}, Location{2, 5}}))
	PrintEditor(&e)

	expected := []string{"0000", "13333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	PrintEditor(&e)
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
	PrintEditor(&e)
	e.Do(SingleDelete(Range{Location{0, 0}, Location{1, 1}}))
	PrintEditor(&e)

	expected := []string{"111", "2222", "3333"}

	expectedCursors := []*Range{
		{Location{0,2},Location{0,3}},
		{Location{1,3},Location{1,4}},
	}

	assert.Equal(t, expected, e.Buffer.Lines)
	assert.Equal(t, expectedCursors, e.Cursors)

	e.Undo()
	PrintEditor(&e)
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
	PrintEditor(&e)
	e.CursorDo(Delete())
	PrintEditor(&e)

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
	PrintEditor(&e)
	e.CursorDo(Delete())
	PrintEditor(&e)

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
	PrintEditor(&e)
	e.CursorDo(Delete())
	PrintEditor(&e)

	expected := []string{"0000", "11113333"}

	assert.Equal(t, expected, e.Buffer.Lines)

	e.Undo()
	assert.Equal(t, linesCopy, e.Buffer.Lines)
}
