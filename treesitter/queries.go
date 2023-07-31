package treesitter

import "embed"

//go:embed queries
var f embed.FS

func GetQuery(language string) ([]byte, error) {
	filename := "queries/" + language + ".scm"
	return f.ReadFile(filename)
}
