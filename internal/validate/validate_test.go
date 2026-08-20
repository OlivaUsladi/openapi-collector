package validate

import (
	"strings"
	"testing"
)

func minimalDoc() map[string]any {
	return map[string]any{
		"openapi": "3.0.3",
		"info":    map[string]any{"title": "T", "version": "1.0.0"},
		"paths":   map[string]any{},
	}
}

func hasError(t *testing.T, errs []error, substr string) {
	for _, e := range errs {
		if strings.Contains(e.Error(), substr) {
			return
		}
	}
	t.Errorf("не найдена ошибка с текстом %q, есть: %v", substr, errs)
}

func TestCheck1(t *testing.T) {
	doc := minimalDoc()
	doc["paths"] = map[string]any{
		"/tasks": map[string]any{
			"get": map[string]any{
				"operationId": "listTasks",
				"responses":   map[string]any{"200": map[string]any{"description": "ok"}},
			},
		},
	}
	errs := Check(doc, nil)
	if len(errs) != 0 {
		t.Fatalf("документ валиден, но получены ошибки: %v", errs)
	}
}

func TestCheck2(t *testing.T) {
	doc := minimalDoc()
	doc["openapi"] = "2.0"
	doc["info"] = map[string]any{"title": ""}
	errs := Check(doc, nil)
	hasError(t, errs, `"openapi" must be a 3.0.x version`)
	hasError(t, errs, `"info.title" is required`)
	hasError(t, errs, `"info.version" is required`)
}

func TestCheck3(t *testing.T) {
	doc := minimalDoc()
	doc["paths"] = map[string]any{
		"/tasks": map[string]any{"get": map[string]any{"operationId": "listTasks"}},
	}
	hasError(t, Check(doc, nil), "operation GET /tasks has no responses")
}

func TestCheck4(t *testing.T) {
	op := func() map[string]any {
		return map[string]any{
			"operationId": "listTasks",
			"responses":   map[string]any{"200": map[string]any{"description": "ok"}},
		}
	}
	doc := minimalDoc()
	doc["paths"] = map[string]any{
		"/tasks":    map[string]any{"get": op()},
		"/archived": map[string]any{"get": op()},
	}
	hasError(t, Check(doc, nil), `duplicate operationId "listTasks"`)
}

func TestCheck5(t *testing.T) {
	doc := minimalDoc()
	doc["paths"] = map[string]any{
		"/tasks/{taskId}": map[string]any{
			"get": map[string]any{
				"responses": map[string]any{"200": map[string]any{"description": "ok"}},
			},
		},
	}
	hasError(t, Check(doc, nil), "path parameter {taskId} of GET /tasks/{taskId} is not described")
}

func TestCheck6(t *testing.T) {
	doc := minimalDoc()
	doc["components"] = map[string]any{
		"parameters": map[string]any{
			"TaskID": map[string]any{"name": "taskId", "in": "path", "required": true},
		},
	}
	doc["paths"] = map[string]any{
		"/tasks/{taskId}": map[string]any{
			"parameters": []any{map[string]any{"$ref": "#/components/parameters/TaskID"}},
			"get": map[string]any{
				"responses": map[string]any{"200": map[string]any{"description": "ok"}},
			},
		},
	}
	errs := Check(doc, nil)
	if len(errs) != 0 {
		t.Fatalf("параметр описан через $ref, ошибок быть не должно: %v", errs)
	}
}

func TestCheck7(t *testing.T) {
	doc := minimalDoc()
	doc["security"] = []any{map[string]any{"apiKey": []any{}}}
	doc["paths"] = map[string]any{
		"/tasks": map[string]any{
			"get": map[string]any{
				"security":  []any{map[string]any{"oauth": []any{}}},
				"responses": map[string]any{"200": map[string]any{"description": "ok"}},
			},
		},
	}
	errs := Check(doc, nil)
	hasError(t, errs, `security scheme "apiKey" used in global security`)
	hasError(t, errs, `security scheme "oauth" used in GET /tasks`)
}
