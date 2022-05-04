package core

import (
	"io"
	"strings"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune([]string{"0000", "1111", "2222", "3333"}))
	e := &Editor{Buffer: b}

	SetCursors(1,1,1,2)(e)
	AsEdit(Insert([]rune{'!'}))(e)

	data, err := io.ReadAll(NewEditorReader(e, 1, 0))

	assert.Nil(t, err)
	expected := []byte("1!111\n2222\n3333\n")

	assert.Equal(t, expected, data)
}

func TestReaderLongLine(t *testing.T) {
	b := NewBuffer()
	line := strings.Repeat("abcdefghijklmnopqrstuvwxyz", 100)
	b.Current = b.Current.Insert(0, ToRune([]string{line}))
	e := &Editor{Buffer: b}

	data, err := io.ReadAll(NewEditorReader(e, 0, 0))

	assert.Nil(t, err)
	expected := []byte(line + "\n")

	assert.Equal(t, expected, data)
}

func TestUnread(t *testing.T) {
	b := NewBuffer()
	b.Current = b.Current.Insert(0, ToRune([]string{"0000", "1111", "2222", "3333"}))
	e := &Editor{Buffer: b}

	SetCursors(1,1,1,2)(e)
	AsEdit(Insert([]rune{'!'}))(e)

	reader := NewEditorReader(e, 3, 0)

	var data []rune
	for i := 0; i < 100; i++ {
		char, _, err := reader.UnreadRune()
		if err != nil {
			break
		}
		data = append(data, char)
	}

	expected := []rune("\n2222\n111!1\n0000")

	assert.Equal(t, expected, data)
}
