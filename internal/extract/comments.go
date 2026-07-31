package extract

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type Comment struct {
	File string
	Line int
	Col  int
	Text string
}

func parseFile(fset *token.FileSet, path string) (*ast.File, error) {
	return parser.ParseFile(fset, path, nil, parser.ParseComments)
}

func FindAPIComments(path string) ([]Comment, error) {
	fset := token.NewFileSet()
	file, err := parseFile(fset, path)
	if err != nil {
		return nil, err
	}

	var result []Comment
	for _, group := range file.Comments {
		if !hasMarker(group) {
			continue
		}
		pos := fset.Position(group.Pos())
		result = append(result, Comment{
			File: path,
			Line: pos.Line,
			Col:  pos.Column,
			Text: group.Text(),
		})
	}
	return result, nil
}

// Начинается с @openapi или просто содержит его?
func hasMarker(group *ast.CommentGroup) bool {
	for _, c := range group.List {
		if strings.Contains(c.Text, "@openapi") {
			return true
		}
	}
	return false
}
