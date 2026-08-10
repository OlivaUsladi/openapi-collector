package spec

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, name, content string) string {
	path := filepath.Join(t.TempDir(), name)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadBaseYAML(t *testing.T) {
	path := writeFile(t, "base.yaml", `
openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas: {}
`)
	doc, err := LoadBase(path)
	if err != nil {
		t.Fatal(err)
	}
	info, ok := doc.Root["info"].(map[string]any)
	if !ok || info["title"] != "Test API" {
		t.Errorf("ожидался title=Test API, получено %v", doc.Root["info"])
	}
}

func TestLoadBaseJSON(t *testing.T) {
	path := writeFile(t, "base.json", `
{
  "openapi": "3.0.3",
  "info": {"title": "Test API", "version": "1.0.0"},
  "paths": {},
  "components": {}
}`)
	doc, err := LoadBase(path)
	if err != nil {
		t.Fatal(err)
	}
	info, ok := doc.Root["info"].(map[string]any)
	if !ok || info["title"] != "Test API" {
		t.Errorf("ожидался title=Test API, получено %v", doc.Root["info"])
	}
}

func TestLoadBaseFail1(t *testing.T) {
	_, err := LoadBase(filepath.Join(t.TempDir(), "no-file.yaml"))
	if err == nil {
		t.Fatal("ожидалась ошибка для несуществующего файла")
	}
}

func TestLoadBaseFail2(t *testing.T) {
	path := writeFile(t, "bad.yaml", "openapi: [pururu")
	_, err := LoadBase(path)
	if err == nil {
		t.Fatal("ожидалась ошибка синтаксиса")
	}
}

func TestLoadBaseFail3(t *testing.T) {
	path := writeFile(t, "no-info.yaml", `
openapi: 3.0.3
paths: {}
components: {}
`)
	_, err := LoadBase(path)
	if err == nil {
		t.Fatalf("ожидалась ошибка про отсутствующее поле info, получено: %v", err)
	}
}

func TestLoadBaseFail4(t *testing.T) {
	path := writeFile(t, "list.yaml", "- a\n- b\n")
	_, err := LoadBase(path)
	if err == nil {
		t.Fatal("ожидалась ошибка про map")
	}
}
