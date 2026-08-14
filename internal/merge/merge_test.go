package merge

import (
	"openapi-collector/internal/model"
	"openapi-collector/internal/spec"
	"strings"
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

func TestMerge2(t *testing.T) {
	res := Merge(baseDoc(t, base), []model.Fragment{
		frag(t, "api/list.go", 3, `
paths:
  /tasks:
    get:
      operationId: listTasks
`),
		frag(t, "api/tasks.go", 3, `
paths:
  /tasks:
    get:
      operationId: getTasks
`),
	})
	if len(res.Conflicts) != 1 {
		t.Fatalf("ожидался 1 конфликт, получено %d: %v", len(res.Conflicts), res.Conflicts)
	}
	c := res.Conflicts[0]
	if c.Kind != "operation" || c.Key != "GET /tasks" {
		t.Errorf("неожиданный конфликт: %+v", c)
	}
	if c.First.File != "api/list.go" || c.Second.File != "api/tasks.go" {
		t.Errorf("неожиданные источники: %+v", c)
	}
	expected := "duplicate operation GET /tasks:\n  first definition: api/list.go:3:1\n  second definition: api/tasks.go:3:1"
	if c.String() != expected {
		t.Errorf("формат конфликта: ожидалось:\n %s \nполучено:\n %s", expected, c.String())
	}
}

func TestMerge3(t *testing.T) {
	res := Merge(baseDoc(t, base), []model.Fragment{
		frag(t, "api/params.go", 3, `
paths:
  /tasks/{taskId}:
    parameters:
      - $ref: "#/components/parameters/TaskID"
`),
		frag(t, "api/get.go", 3, `
paths:
  /tasks/{taskId}:
    get:
      operationId: getTask
`),
	})
	if len(res.Conflicts) != 0 || len(res.Errors) != 0 {
		t.Fatalf("неожиданные проблемы: %v %v", res.Conflicts, res.Errors)
	}
	item := res.Doc["paths"].(map[string]any)["/tasks/{taskId}"].(map[string]any)
	if item["parameters"] == nil || item["get"] == nil {
		t.Errorf("ожидались parameters и get вместе, получено %v", item)
	}
}

func TestMerge4(t *testing.T) {
	paramFragment := `
paths:
  /tasks:
    parameters:
      - name: page
        in: query
        schema:
          type: integer
`
	res := Merge(baseDoc(t, base), []model.Fragment{
		frag(t, "a.go", 3, paramFragment),
		frag(t, "b.go", 3, paramFragment),
	})
	if len(res.Conflicts) != 0 {
		t.Fatalf("идентичный параметр должен быть дублем, а не конфликтом: %v", res.Conflicts)
	}
	item := res.Doc["paths"].(map[string]any)["/tasks"].(map[string]any)
	if params := item["parameters"].([]any); len(params) != 1 {
		t.Errorf("ожидался 1 параметр после дедупликации, получено %d", len(params))
	}
}

func TestMerge5(t *testing.T) {
	res := Merge(baseDoc(t, base), []model.Fragment{
		frag(t, "a.go", 3, `
paths:
  /tasks:
    parameters:
      - name: page
        in: query
        schema:
          type: integer
`),
		frag(t, "b.go", 3, `
paths:
  /tasks:
    parameters:
      - name: page
        in: query
        schema:
          type: string
`),
	})
	if len(res.Conflicts) != 1 || res.Conflicts[0].Kind != "parameter" {
		t.Fatalf("ожидался конфликт параметра, получено %v", res.Conflicts)
	}
}

func TestMerge6(t *testing.T) {
	res := Merge(baseDoc(t, base), []model.Fragment{
		frag(t, "a.go", 3, `
components:
  schemas:
    Task:
      type: object
`),
		frag(t, "b.go", 3, `
components:
  responses:
    Task:
      description: Task response
`),
	})
	if len(res.Conflicts) != 0 {
		t.Fatalf("одинаковые имена в разных разделах не конфликтуют: %v", res.Conflicts)
	}
}

func TestMerge7(t *testing.T) {
	res := Merge(baseDoc(t, base), []model.Fragment{
		frag(t, "domain/task.go", 3, `
components:
  schemas:
    Task:
      type: object
      properties:
        id:
          type: integer
`),
		frag(t, "api/task.go", 3, `
components:
  schemas:
    Task:
      type: object
      properties:
        title:
          type: string
`),
	})
	if len(res.Conflicts) != 1 {
		t.Fatalf("ожидался 1 конфликт, получено %v", res.Conflicts)
	}
	c := res.Conflicts[0]
	if c.Key != "components.schemas.Task" {
		t.Errorf("неожиданный ключ конфликта: %s", c.Key)
	}
	if c.First.File != "api/task.go" || c.Second.File != "domain/task.go" {
		t.Errorf("неожиданные источники: %+v", c)
	}
}

func TestMerge8(t *testing.T) {
	res := Merge(baseDoc(t, `
openapi: 3.0.3
info:
  title: T
  version: 1.0.0
paths: {}
components:
  schemas:
    Task:
      type: object
`), []model.Fragment{
		frag(t, "api/task.go", 3, `
components:
  schemas:
    Task:
      type: string
`),
	})
	if len(res.Conflicts) != 1 {
		t.Fatalf("ожидался конфликт с базой, получено %v", res.Conflicts)
	}
	if res.Conflicts[0].First.File != "spec/base.yaml" {
		t.Errorf("первым источником должна быть база: %+v", res.Conflicts[0])
	}
}

func TestMerge9(t *testing.T) {
	res := Merge(baseDoc(t, base), []model.Fragment{
		frag(t, "api/bad.go", 3, `
info:
  title: Tratata
`),
	})
	if len(res.Errors) != 1 || !strings.Contains(res.Errors[0].Error(), `top-level field "info"`) {
		t.Fatalf("ожидалась ошибка про запрещённое поле info: %v", res.Errors)
	}
}

func TestMerge10(t *testing.T) {
	res := Merge(baseDoc(t, base), []model.Fragment{
		frag(t, "api/bad.go", 3, `
definitions:
  Task: {}
`),
	})
	if len(res.Errors) != 1 || !strings.Contains(res.Errors[0].Error(), "неизвестное поле") {
		t.Fatalf("ожидалась ошибка про неизвестное поле: %v", res.Errors)
	}
}

func TestMerge11(t *testing.T) {
	sameTag := `
tags:
  - name: tasks
    description: Operations with tasks
`
	res := Merge(baseDoc(t, base), []model.Fragment{
		frag(t, "a.go", 3, sameTag),
		frag(t, "b.go", 3, sameTag),
		frag(t, "c.go", 3, `
tags:
  - name: users
    description: Operations with users
`),
	})
	if len(res.Conflicts) != 0 {
		t.Fatalf("неожиданные конфликты: %v", res.Conflicts)
	}
	if tags := res.Doc["tags"].([]any); len(tags) != 2 {
		t.Errorf("ожидалось 2 тега, получено %d", len(tags))
	}
}

func TestMerge12(t *testing.T) {
	res := Merge(baseDoc(t, base), []model.Fragment{
		frag(t, "a.go", 3, `
tags:
  - name: tasks
    description: One
`),
		frag(t, "b.go", 3, `
tags:
  - name: tasks
    description: Another
`),
	})
	if len(res.Conflicts) != 1 || res.Conflicts[0].Kind != "tag" {
		t.Fatalf("ожидался конфликт тега, получено %v", res.Conflicts)
	}
}

func TestMerge13(t *testing.T) {
	res := Merge(baseDoc(t, base), []model.Fragment{
		frag(t, "a.go", 3, `
paths:
  no-slash:
    get:
      operationId: x
  /double//slash:
    get:
      operationId: y
  /empty: {}
`),
	})
	if len(res.Errors) != 3 {
		t.Fatalf("ожидались 3 ошибки путей, получено %d: %v", len(res.Errors), res.Errors)
	}
}

func TestMerge14(t *testing.T) {
	f1 := frag(t, "a.go", 3, `
paths:
  /tasks:
    get:
      operationId: listTasks
`)
	f2 := frag(t, "b.go", 3, `
components:
  schemas:
    Task:
      type: object
`)
	f3 := frag(t, "c.go", 3, `
paths:
  /tasks:
    post:
      operationId: createTask
`)

	res1 := Merge(baseDoc(t, base), []model.Fragment{f1, f2, f3})
	res2 := Merge(baseDoc(t, base), []model.Fragment{f3, f1, f2})

	if !deepEqual(any(res1.Doc), any(res2.Doc)) {
		t.Error("результат merge зависит от порядка фрагментов")
	}
}

func TestMerge15(t *testing.T) {
	res := Merge(baseDoc(t, base), []model.Fragment{
		frag(t, "a.go", 3, "paths:\n  /x:\n    get:\n      operationId: a\n"),
		frag(t, "b.go", 3, "paths:\n  /x:\n    get:\n      operationId: b\n"),
		frag(t, "c.go", 3, "components:\n  schemas:\n    T:\n      type: object\n"),
		frag(t, "d.go", 3, "components:\n  schemas:\n    T:\n      type: string\n"),
	})
	if len(res.Conflicts) != 2 {
		t.Fatalf("ожидались 2 конфликта за один запуск, получено %d: %v",
			len(res.Conflicts), res.Conflicts)
	}

	if res.Conflicts[0].Kind != "component" || res.Conflicts[1].Kind != "operation" {
		t.Errorf("неожиданный порядок конфликтов: %v", res.Conflicts)
	}
}

func TestMerge16(t *testing.T) {
	base := baseDoc(t, base)
	Merge(base, []model.Fragment{
		frag(t, "a.go", 3, "components:\n  schemas:\n    T:\n      type: object\n"),
	})
	schemas := base.Root["components"].(map[string]any)["schemas"].(map[string]any)
	if len(schemas) != 0 {
		t.Error("merge не должен менять исходную базовую спецификацию")
	}
}
