package treesitter

import (
	sitter "github.com/smacker/go-tree-sitter"
	"fmt"
)

import (
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/rust"
)

type Language struct {
	name string
	sitter *sitter.Language
	query []byte
}

func GetLanguage(name string) (*Language, error) {
	query, err := GetQuery(name)
	if err != nil {
		return nil, err
	}
	switch name {
	case "javascript":
		return &Language{
			name: "javascript",
			sitter: javascript.GetLanguage(),
			query: query,
		}, nil
	case "c":
		return &Language{
			name: "c",
			sitter: c.GetLanguage(),
			query: query,
		}, nil
	case "rust":
		return &Language{
			name: "rust",
			sitter: rust.GetLanguage(),
			query: query,
		}, nil
	default:
		return nil, fmt.Errorf("Unknown language: %s", name)
	}
}
