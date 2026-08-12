package merge

import (
	"fmt"
	"openapi-collector/internal/model"
	"strings"
)

var httpMethods = []string{"get", "post", "put", "patch", "delete", "options", "head", "trace"}

var pathLevelFields = map[string]bool{
	"summary": true, "description": true, "servers": true, "parameters": true,
}

func isHTTPMethod(field string) bool {
	lower := strings.ToLower(field)
	for _, m := range httpMethods {
		if m == lower {
			return true
		}
	}
	return false
}

// вливает раздел paths одного фрагмента
func (m *merger) mergePaths(value any, origin model.Origin) {
	paths, ok := value.(map[string]any)
	if !ok {
		m.errorf("%s:%d: фрагмент поля \"paths\" должен быть map",
			origin.File, origin.Line)
		return
	}

	docPaths, ok := m.res.Doc["paths"].(map[string]any)
	if !ok {
		docPaths = map[string]any{}
		m.res.Doc["paths"] = docPaths
	}

	for path, rawItem := range paths {
		err := checkPath(path)
		if err != nil {
			m.errorf("%s:%d: %v", origin.File, origin.Line, err)
			continue
		}
		item, ok := rawItem.(map[string]any)
		if !ok {
			m.errorf("%s:%d: элемент path %q должен быть map", origin.File, origin.Line, path)
			continue
		}
		if len(item) == 0 {
			m.errorf("%s:%d: path %q не содержит операций и полей",
				origin.File, origin.Line, path)
			continue
		}

		docItem, ok := docPaths[path].(map[string]any)
		if !ok {
			docItem = map[string]any{}
			docPaths[path] = docItem
		}

		for field, fieldValue := range item {
			if isHTTPMethod(field) {
				m.mergeOperation(docItem, path, field, fieldValue, origin)
				continue
			}

			if field == "parameters" {
				m.mergePathParameters(docItem, path, fieldValue, origin)
				continue
			}

			if pathLevelFields[field] {
				m.mergePathField(docItem, path, field, fieldValue, origin)
				continue
			}

			if strings.HasPrefix(field, "x-") {
				m.mergePathField(docItem, path, field, fieldValue, origin)
				continue
			}

			m.errorf("%s:%d: неизвестное поле %q в path %q", origin.File, origin.Line, field, path)
		}
	}
}

// проверяет ведущий / и отсутствие //
func checkPath(path string) error {
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path %q должен начинаться с /", path)
	}
	if strings.Contains(path, "//") {
		return fmt.Errorf("path %q содержит //", path)
	}
	return nil
}

// добавляет HTTP-операцию
func (m *merger) mergeOperation(docItem map[string]any, path, method string, value any, origin model.Origin) {
	lower := strings.ToLower(method)
	key := operationKey(lower, path)

	f, ex := m.res.Owners[key]
	if ex {
		m.conflict("operation", strings.ToUpper(lower)+" "+path, f, origin)
		return
	}
	docItem[lower] = value
	m.res.Owners[key] = origin
}

// объединяет общие параметры пути
func (m *merger) mergePathParameters(docItem map[string]any, path string, value any, origin model.Origin) {
	params, ok := value.([]any)
	if !ok {
		m.errorf("%s:%d: поле \"parameters\" пути %q должно быть массивом", origin.File, origin.Line, path)
		return
	}

	existing, _ := docItem["parameters"].([]any)

	for _, rawParam := range params {
		param, ok := rawParam.(map[string]any)
		if !ok {
			m.errorf("%s:%d: параметр пути %q должен быть map", origin.File, origin.Line, path)
			continue
		}
		identity, ok := parameterIdentity(param)
		if !ok {
			m.errorf("%s:%d: параметр пути %q должен содержать \"name\" и \"in\" либо \"$ref\"",
				origin.File, origin.Line, path)
			continue
		}

		key := "param:" + path + " " + identity
		first, exists := m.res.Owners[key]
		if exists {
			var stored map[string]any
			for _, e := range existing {
				em, ok := e.(map[string]any)
				if ok {
					if id, _ := parameterIdentity(em); id == identity {
						stored = em
						break
					}
				}
			}
			if !deepEqual(any(stored), any(param)) {
				m.conflict("parameter", identity+" "+path, first, origin)
			}
			continue
		}

		existing = append(existing, param)
		m.res.Owners[key] = origin
	}

	if len(existing) > 0 {
		docItem["parameters"] = existing
	}
}

// идентичность параметра
// $ref или name + in
func parameterIdentity(param map[string]any) (string, bool) {
	ref, ok := param["$ref"].(string)
	if ok {
		return "$ref " + ref, true
	}
	name, okName := param["name"].(string)
	in, okIn := param["in"].(string)
	if !okName || !okIn {
		return "", false
	}
	return "name " + name + " in " + in, true
}

// добавляет одиночное поле уровня пути
func (m *merger) mergePathField(docItem map[string]any, path, field string, value any, origin model.Origin) {
	key := "pathfield:" + path + " " + field

	first, exists := m.res.Owners[key]
	if exists {
		if !deepEqual(docItem[field], value) {
			m.conflict("path-level field", field+" "+path, first, origin)
		}
		return
	}
	docItem[field] = value
	m.res.Owners[key] = origin
}
