#!/bin/bash

cd "$(dirname "$0")"

languages=("javascript" "c" "rust")

for language in ${languages[@]}; do
	curl "https://raw.githubusercontent.com/tree-sitter/tree-sitter-${language}/master/queries/highlights.scm" \
		> "$language.scm"
done
