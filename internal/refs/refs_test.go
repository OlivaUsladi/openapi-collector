package refs

import (
	"openapi-collector/internal/model"
	"strings"
	"testing"
)

func doc() map[string]any {
	return map[string]any{
		"paths": map[string]any{
			"/tasks": map[string]any{
				"get": map[string]any{
					"responses": map[string]any{
						"200": map[string]any{
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{"$ref": "#/components/schemas/Task"},
								},
							},
						},
					},
				},
			},
		},
		"components": map[string]any{
			"schemas": map[string]any{
				"Task": map[string]any{"type": "object"},
			},
		},
	}
}

func TestCheck1(t *testing.T) {
	errs, warnings := Check(doc(), nil)
	if len(errs) != 0 {
		t.Fatalf("не ожидались ошибки, получено: %v", errs)
	}
	if len(warnings) != 0 {
		t.Fatalf("не ожидались предупреждения, получено: %v", warnings)
	}
}

func TestCheck2(t *testing.T) {
	d := doc()
	owners := map[string]model.Origin{
		"op:get /tasks": {File: "api/tasks.go", Line: 5},
	}
	schema := d["paths"].(map[string]any)["/tasks"].(map[string]any)["get"].(map[string]any)
	schema["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"] = map[string]any{
		"$ref": "#/components/schemas/Missing",
	}

	errs, _ := Check(d, owners)
	if len(errs) != 1 {
		t.Fatalf("ожидалась 1 ошибка, получено: %v", errs)
	}
	msg := errs[0].Error()
	if !strings.Contains(msg, "unresolved reference: #/components/schemas/Missing") {
		t.Errorf("нет текста unresolved reference: %s", msg)
	}
	if !strings.Contains(msg, "api/tasks.go:5") {
		t.Errorf("нет источника фрагмента: %s", msg)
	}
}

func TestCheck3(t *testing.T) {
	d := doc()
	d["components"].(map[string]any)["schemas"].(map[string]any)["Task"] = map[string]any{
		"$ref": "other.yaml#/components/schemas/Task",
	}
	errs, _ := Check(d, nil)
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "external reference") {
		t.Fatalf("ожидалась ошибка external reference, получено: %v", errs)
	}
}

func TestCheck4(t *testing.T) {
	d := doc()
	d["components"].(map[string]any)["schemas"].(map[string]any)["Orphan"] = map[string]any{"type": "string"}
	errs, warnings := Check(d, nil)
	if len(errs) != 0 {
		t.Fatalf("не ожидались ошибки: %v", errs)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "components.schemas.Orphan") {
		t.Fatalf("ожидалось предупреждение об Orphan, получено: %v", warnings)
	}
}

func TestCheck5(t *testing.T) {
	d := map[string]any{
		"components": map[string]any{
			"schemas": map[string]any{
				"A": map[string]any{"properties": map[string]any{"b": map[string]any{"$ref": "#/components/schemas/B"}}},
				"B": map[string]any{"properties": map[string]any{"a": map[string]any{"$ref": "#/components/schemas/A"}}},
			},
		},
	}
	errs, _ := Check(d, nil)
	if len(errs) != 0 {
		t.Fatalf("циклические ссылки не должны давать ошибок: %v", errs)
	}
}
