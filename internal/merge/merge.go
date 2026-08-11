package merge

import (
	"fmt"
	"openapi-collector/internal/model"
	"openapi-collector/internal/spec"
	"sort"
	"strings"
)

var baseFields = []string{"openapi", "info", "servers", "externalDocs", "security"}

type Conflict struct {
	Kind   string
	Key    string
	First  model.Origin
	Second model.Origin
}

type Result struct {
	Doc       map[string]any
	Owners    map[string]model.Origin
	Conflicts []Conflict
	Errors    []error
}

type merger struct {
	res *Result
}

/*
Ожидаемая ошибка:
duplicate component components.schemas.Task:
  first definition: domain/task.go:3:5
  second definition: api/task.go:3:5
*/

func (c Conflict) String() string {
	return fmt.Sprintf("duplicate %s %s:\n  first definition: %s\n  second definition: %s",
		c.Kind, c.Key, formatOrigin(c.First), formatOrigin(c.Second))
}

func formatOrigin(o model.Origin) string {
	if o.Line == 0 {
		return o.File
	}
	return fmt.Sprintf("%s:%d:%d", o.File, o.Line, o.Column)
}

// объединяет базовую спецификацию со всеми фрагментами
func Merge(base *spec.Document, fragments []model.Fragment) *Result {
	res := &Result{
		Doc:    copyValue(base.Root).(map[string]any),
		Owners: map[string]model.Origin{},
	}

	m := &merger{res: res}

	baseOrigin := model.Origin{File: base.Path}
	m.registerBase(baseOrigin)

	sorted := make([]model.Fragment, len(fragments))
	copy(sorted, fragments)
	sort.Slice(sorted, func(i, j int) bool {
		a, b := sorted[i].Origin, sorted[j].Origin
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Column < b.Column
	})

	for _, frag := range sorted {
		m.mergeFragment(frag)
	}

	m.sortIssues()
	return m.res
}

// записывает владельцев для операций, компонентов и тегов
func (m *merger) registerBase(origin model.Origin) {
	paths, ok := m.res.Doc["paths"].(map[string]any)
	if ok {
		for path, item := range paths {
			pathItem, ok := item.(map[string]any)
			if !ok {
				continue
			}
			for field := range pathItem {
				// здесь будет проверка isHTTPMethod(field) различаются ли HTTP методы и path поля
				_ = field
				m.res.Owners[operationKey(field, path)] = origin
			}
		}
	}
	components, ok := m.res.Doc["components"].(map[string]any)
	if ok {
		for section, names := range components {
			sectionMap, ok := names.(map[string]any)
			if !ok {
				continue
			}
			for name := range sectionMap {
				m.res.Owners[componentKey(section, name)] = origin
			}
		}
	}
	tags, ok := m.res.Doc["tags"].([]any)
	if ok {
		for _, t := range tags {
			tag, ok := t.(map[string]any)
			if ok {
				name, ok := tag["name"].(string)
				if ok {
					m.res.Owners[tagKey(name)] = origin
				}
			}
		}
	}
}

/*
Пример запрещённого поля:
api/tasks.go:15: fragment must not redefine top-level field "info"
*/

// вливает один фрагмент в итоговый документ
func (m *merger) mergeFragment(frag model.Fragment) {
	ok := true
	for key := range frag.Doc {
		if isBase(key) {
			m.errorf("%s:%d: fragment must not redefine top-level field %q",
				frag.Origin.File, frag.Origin.Line, key)
			ok = false
		} else if key != "paths" && key != "components" && key != "tags" &&
			!strings.HasPrefix(key, "x-") {
			m.errorf("%s:%d: неизвестное поле %q во фрагменте",
				frag.Origin.File, frag.Origin.Line, key)
			ok = false
		}
	}
	if !ok {
		return
	}

	_, has := frag.Doc["paths"]
	if has {
		// здесь будет m.mergePaths(paths, frag.Origin)
	}
	_, has = frag.Doc["components"]
	if has {
		// здесь будет m.mergeComponents(components, frag.Origin)
	}
	_, has = frag.Doc["tags"]
	if has {
		// здесь будет m.mergeTags(tags, frag.Origin)
	}
	// здесь будет перенос расширений x-* верхнего уровня фрагмента в итоговый документ
}

// запрещено ли поле верхнего уровня во фрагменте
func isBase(field string) bool {
	for _, f := range baseFields {
		if f == field {
			return true
		}
	}
	return false
}

// добавляет ошибку в результат
func (m *merger) errorf(format string, args ...any) {
	m.res.Errors = append(m.res.Errors, fmt.Errorf(format, args...))
}

// добавляет конфликт в результат
func (m *merger) conflict(kind, key string, first, second model.Origin) {
	m.res.Conflicts = append(m.res.Conflicts, Conflict{
		Kind: kind, Key: key, First: first, Second: second,
	})
}

// приводит конфликты и ошибки к порядку вывода
func (m *merger) sortIssues() {
	sort.Slice(m.res.Conflicts, func(i, j int) bool {
		a := m.res.Conflicts[i]
		b := m.res.Conflicts[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Key < b.Key
	})
	sort.Slice(m.res.Errors, func(i, j int) bool {
		return m.res.Errors[i].Error() < m.res.Errors[j].Error()
	})
}

// ключ операции в карте владельцев, например "op:get /tasks".
func operationKey(method, path string) string {
	return "op:" + strings.ToLower(method) + " " + path
}

// ключ компонента, например "comp:schemas.Task".
func componentKey(section, name string) string {
	return "comp:" + section + "." + name
}

// ключ тега, например "tag:tasks".
func tagKey(name string) string {
	return "tag:" + name
}
