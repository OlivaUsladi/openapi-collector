package output

import (
	"strings"
	"testing"
)

func sampleDoc() map[string]any {
	return map[string]any{
		"components": map[string]any{
			"schemas": map[string]any{
				"Task": map[string]any{
					"type":       "object",
					"properties": map[string]any{"id": map[string]any{"type": "string"}},
				},
			},
		},
		"openapi": "3.0.3",
		"info":    map[string]any{"version": "1.0.0", "title": "Задачи <и> дела"},
		"paths": map[string]any{
			"/tasks": map[string]any{
				"post": map[string]any{"responses": map[string]any{"201": map[string]any{"description": "created"}}},
				"get":  map[string]any{"responses": map[string]any{"200": map[string]any{"description": "ok"}}},
			},
		},
	}
}

func TestMarshal1(t *testing.T) {
	first, err := MarshalYAML(sampleDoc())
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		next, err := MarshalYAML(sampleDoc())
		if err != nil {
			t.Fatal(err)
		}
		if string(next) != string(first) {
			t.Fatalf("повторный вызов дал другой результат:\n %s\n %s", first, next)
		}
	}
}

func TestMarshal2(t *testing.T) {
	data, err := MarshalYAML(sampleDoc())
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	openapiPos := strings.Index(text, "openapi:")
	infoPos := strings.Index(text, "info:")
	pathsPos := strings.Index(text, "paths:")
	componentsPos := strings.Index(text, "components:")
	if !(openapiPos < infoPos && infoPos < pathsPos && pathsPos < componentsPos) {
		t.Errorf("неверный порядок верхнего уровня:\n %s", text)
	}
	getPos := strings.Index(text, "  get:")
	postPos := strings.Index(text, "  post:")
	if !(getPos >= 0 && postPos >= 0 && getPos < postPos) {
		t.Errorf("get должен идти раньше post:\n %s", text)
	}
}

func TestMarshal3(t *testing.T) {
	data, err := MarshalYAML(sampleDoc())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"200":`) {
		t.Errorf("код ответа должен остаться строкой в кавычках:\n %s", data)
	}
}

func TestMarshal4(t *testing.T) {
	data, err := MarshalYAML(sampleDoc())
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.HasSuffix(text, "\n") || strings.HasSuffix(text, "\n\n") {
		t.Errorf("файл должен заканчиваться одним переводом строки")
	}
	if strings.Contains(text, "\t") || strings.Contains(text, "\r") {
		t.Errorf("в YAML не должно быть TAB и \\r")
	}
}
