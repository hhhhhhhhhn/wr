package treesitter

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

const query = `
; Special identifiers
;--------------------

([
    (identifier)
    (shorthand_property_identifier)
    (shorthand_property_identifier_pattern)
 ] @constant
 (#match? @constant "^[A-Z_][A-Z\\d_]+$"))


((identifier) @constructor
 (#match? @constructor "^[A-Z]"))

((identifier) @variable.builtin
 (#match? @variable.builtin "^(arguments|module|console|window|document)$")
 (#is-not? local))

((identifier) @function.builtin
 (#eq? @function.builtin "require")
 (#is-not? local))

; Function and method definitions
;--------------------------------

(function
  name: (identifier) @function)
(function_declaration
  name: (identifier) @function)
(method_definition
  name: (property_identifier) @function.method)

(pair
  key: (property_identifier) @function.method
  value: [(function) (arrow_function)])

(assignment_expression
  left: (member_expression
    property: (property_identifier) @function.method)
  right: [(function) (arrow_function)])

(variable_declarator
  name: (identifier) @function
  value: [(function) (arrow_function)])

(assignment_expression
  left: (identifier) @function
  right: [(function) (arrow_function)])

; Function and method calls
;--------------------------
(call_expression
  function: (identifier) @function)

(call_expression
  function: (member_expression
    property: (property_identifier) @function.method))

; Variables
;----------

(identifier) @variable

; Properties
;-----------

(property_identifier) @property

; Literals
;---------

(this) @variable.builtin
(super) @variable.builtin

[
  (true)
  (false)
  (null)
  (undefined)
] @constant.builtin

(comment) @comment

[
  (string)
  (template_string)
] @string

(regex) @string.special
(number) @number

; Tokens
;-------

(template_substitution
  "${" @punctuation.special
  "}" @punctuation.special) @embedded

[
  ";"
;  (optional_chain)
  "."
  ","
] @punctuation.delimiter

[
  "-"
  "--"
  "-="
  "+"
  "++"
  "+="
  "*"
  "*="
  "**"
  "**="
  "/"
  "/="
  "%"
  "%="
  "<"
  "<="
  "<<"
  "<<="
  "="
  "=="
  "==="
  "!"
  "!="
  "!=="
  "=>"
  ">"
  ">="
  ">>"
  ">>="
  ">>>"
  ">>>="
  "~"
  "^"
  "&"
  "|"
  "^="
  "&="
  "|="
  "&&"
  "||"
  "??"
  "&&="
  "||="
  "??="
] @operator

[
  "("
  ")"
  "["
  "]"
  "{"
  "}"
]  @punctuation.bracket

[
  "as"
  "async"
  "await"
  "break"
  "case"
  "catch"
  "class"
  "const"
  "continue"
  "debugger"
  "default"
  "delete"
  "do"
  "else"
  "export"
  "extends"
  "finally"
  "for"
  "from"
  "function"
  "get"
  "if"
  "import"
  "in"
  "instanceof"
  "let"
  "new"
  "of"
  "return"
  "set"
  "static"
  "switch"
  "target"
  "throw"
  "try"
  "typeof"
  "var"
  "void"
  "while"
  "with"
  "yield"
] @keyword
`

type Highlighter struct {
	query       *sitter.Query
	queryCursor *sitter.QueryCursor
	parser      *sitter.Parser
	root        *sitter.Tree
}

var input = sitter.Input{
	Read: func(offset uint32, position sitter.Point) []byte {
		return []byte{}
	},
	Encoding: sitter.InputEncodingUTF8,
}

func NewHighlighter() *Highlighter {
	query, _ := sitter.NewQuery([]byte(query), javascript.GetLanguage())
	return &Highlighter{
		query:       query,
		queryCursor: sitter.NewQueryCursor(),
		parser:      sitter.NewParser(),
	}
}

func (h *Highlighter) Parse(content []byte) {
	h.root = h.parser.Parse(nil, content)
}

// row is 0 indexed
func (h *Highlighter) ChangeLine(row uint32, startByte uint32, oldLength uint32, newLength uint32, newContent []byte) {
	h.root.Edit(sitter.EditInput{
		StartIndex: startByte,
		OldEndIndex: startByte + oldLength,
		NewEndIndex: startByte + newLength,
		StartPoint: sitter.Point {
			Row: row + 1,
			Column: startByte,
		},
		OldEndPoint: sitter.Point {
			Row: row + 1,
			Column: startByte + oldLength,
		},
		NewEndPoint: sitter.Point {
			Row: row + 1,
			Column: startByte + newLength,
		},
	})
}

const code = `
function displayClosure() {
	var count = "a";
	return function () {
		return ++count;
	};
}
var inc = displayClosure();
inc(); // devuelve 1
inc(); // devuelve 2
inc(); // devuelve 3
`

const code2 = `
function displayClosure() {
	var count = "a";
	return function () {
		return count++;
	};
}
var inc = displayClosure();
inc(); // devuelve 1
inc(); // devuelve 2
inc(); // devuelve 3
`


func main() {
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
	query, e := sitter.NewQuery([]byte(query), javascript.GetLanguage())

	fmt.Println(e)
	queryCursor := sitter.NewQueryCursor()

	code := []byte(code)

	tree := parser.Parse(nil, code)

	tree.Edit(sitter.EditInput{
		StartIndex: 75,
		OldEndIndex: 82,
		NewEndIndex: 82,
		StartPoint: sitter.Point {
			Row: 4,
			Column: 9,
		},
		OldEndPoint: sitter.Point {
			Row: 4,
			Column: 16,
		},
		NewEndPoint: sitter.Point {
			Row: 4,
			Column: 16,
		},
	})

	tree = parser.Parse(tree, []byte(code2))

	node := tree.RootNode()

	queryCursor.Exec(query, node)

	var captures []sitter.QueryCapture

	for true {
		match, ok := queryCursor.NextMatch()
		if !ok {
			break
		}
		for _, capture := range match.Captures {
			if len(captures) > 0 {
				lastCapture := captures[len(captures)-1]
				if intersects(capture, lastCapture) &&
					len(query.CaptureNameForId(capture.Index)) >= len(query.CaptureNameForId(lastCapture.Index)) {
						captures[len(captures)-1] = capture
						continue
				}
			}

			captures = append(captures, capture)
		}
	}
	for _, capture := range captures {
		fmt.Println(query.CaptureNameForId(capture.Index),"\t\t", capture.Node, capture.Node.StartPoint(), capture.Node.EndPoint())
	}
	fmt.Println()
}

// Assuming a is after b
func intersects(a, b sitter.QueryCapture) bool {
	return a.Node.StartByte() < b.Node.EndByte()
}
