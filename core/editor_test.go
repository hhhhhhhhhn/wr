package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSortCursors(t *testing.T) {
	cursors := []*Range{
		{Location{2, 1}, Location{2, 2}},
		{Location{1, 1}, Location{1, 2}},
		{Location{1, 2}, Location{1, 3}},
	}

	sorted := SortCursors(cursors)

	expected := []*Range{
		{Location{1, 1}, Location{1, 2}},
		{Location{1, 2}, Location{1, 3}},
		{Location{2, 1}, Location{2, 2}},
	}

	assert.Equal(t, expected, sorted)
	assert.NotEqual(t, expected, cursors)
}
