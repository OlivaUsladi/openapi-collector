package merge

import (
	"openapi-collector/internal/model"
	"openapi-collector/internal/spec"
	"testing"

	"gopkg.in/yaml.v3"
)

func baseDoc(t *testing.T, text string) *spec.Document {
	var parsed any
	err := yaml.Unmarshal([]byte(text), &parsed)
	if err != nil {
		t.Fatal(err)
	}
	root, ok := parsed.(map[string]any)
	if !ok {
		t.Fatal("база в тесте должна быть map")
	}
	return &spec.Document{Root: root, Path: "spec/base.yaml"}
}

func frag(t *testing.T, file string, line int, text string) model.Fragment {
	var parsed any
	err := yaml.Unmarshal([]byte(text), &parsed)
	if err != nil {
		t.Fatal(err)
	}
	doc, ok := parsed.(map[string]any)
	if !ok {
		t.Fatal("фрагмент в тесте должен быть map")
	}
	return model.Fragment{
		Origin: model.Origin{File: file, Line: line, Column: 1},
		Doc:    doc,
		Raw:    text,
	}
}

const base = `
openapi: 3.0.3
info:
  title: T
  version: 1.0.0
paths: {}
components:
  schemas: {}
`

func TestMerge1(t *testing.T) {
	res := Merge(baseDoc(t, base), []model.Fragment{
		frag(t, "api/list.go", 3, `
paths:
  /tasks:
    get:
      operationId: listTasks
`),
		frag(t, "api/create.go", 3, `
paths:
  /tasks:
    post:
      operationId: createTask
`),
	})
	if len(res.Conflicts) != 0 || len(res.Errors) != 0 {
		t.Fatalf("неожиданные проблемы: %v %v", res.Conflicts, res.Errors)
	}
	item := res.Doc["paths"].(map[string]any)["/tasks"].(map[string]any)
	if item["get"] == nil || item["post"] == nil {
		t.Errorf("ожидались обе операции get и post, получено %v", item)
	}
}
