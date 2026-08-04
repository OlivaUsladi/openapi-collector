package extract

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"openapi-collector/internal/model"
	"strings"
)

type rawFragment struct {
	origin model.Origin
	lines  []string
}

// разбирает исходный текст одного Go-файла и возвращает все найденные @openapi-фрагменты
func FindApiComment(filename, path string, src []byte) ([]model.Fragment, error) {

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("парсинг %s: %w", filename, err)
	}

	var fragments []model.Fragment
	for _, group := range file.Comments {
		pos := fset.Position(group.Pos())
		origin := model.Origin{filename, pos.Line, pos.Column}

		raw, found, err := rawFragmentBuild(group, origin)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}

		frag, err := buildFragment(raw)
		if err != nil {
			return nil, err
		}
		fragments = append(fragments, frag)
	}
	return fragments, nil
}

// превращает группу комментариев в rawFragment
func rawFragmentBuild(group *ast.CommentGroup, origin model.Origin) (rawFragment, bool, error) {
	var lines []string
	for _, c := range group.List {
		lines = append(lines, commentLines(c.Text)...)
	}
	mIdx := -1
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		if trim == "@openapi" {
			mIdx = i
		}
		break
	}
	if mIdx < 0 {
		return rawFragment{}, false, nil
	}

	yLines := lines[mIdx+1:]
	d, err := trimCom(yLines, origin)
	if err != nil {
		return rawFragment{}, false, err
	}
	return rawFragment{origin, d}, true, nil
}

// возвращает строки текста одного комментария без префиксов
func commentLines(text string) []string {
	if strings.HasPrefix(text, "//") {
		line := strings.TrimPrefix(text, "//")
		line = strings.TrimPrefix(line, " ")
		return []string{line}
	}

	text = strings.TrimPrefix(text, "/*")
	text = strings.TrimSuffix(text, "*/")
	lines := strings.Split(text, "\n")

	if hasFirstStars(lines) {
		for i := 1; i < len(lines); i++ {
			trimmed := strings.TrimLeft(lines[i], " \t")
			trimmed = strings.TrimPrefix(trimmed, "*")
			trimmed = strings.TrimPrefix(trimmed, " ")
			lines[i] = trimmed
		}
	}
	return lines
}

// есть ли звёздоски в начале
func hasFirstStars(lines []string) bool {
	found := false
	for _, line := range lines[1:] {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "*") {
			return false
		}
		found = true
	}
	return found
}

// убирает общий отступ по первой непустой строке
func trimCom(lines []string, origin model.Origin) ([]string, error) {
	indent := ""
	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		trimmed := strings.TrimLeft(line, " \t")
		indent = line[:len(line)-len(trimmed)]
		start = i
		break
	}
	if start < 0 {
		return nil, fmt.Errorf("%s:%d:%d: пустой @openapi, нет YAML после маркера",
			origin.File, origin.Line, origin.Column)
	}

	var out []string
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			out = append(out, "")
			continue
		}
		cut := strings.TrimPrefix(line, indent)
		if strings.ContainsAny(indentOf(cut), "\t") || strings.Contains(indent, "\t") {
			return nil, fmt.Errorf("%s:%d: ошибка табуляции %d",
				origin.File, origin.Line+start, i+1)
		}
		out = append(out, cut)
	}
	return out, nil
}

// возвращает ведущие пробельные символы строки.
func indentOf(line string) string {
	return line[:len(line)-len(strings.TrimLeft(line, " \t"))]
}
