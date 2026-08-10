package spec

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Document struct {
	Root map[string]any
	Path string
}

// читает базовую спецификацию из YAML или JSON файла
func LoadBase(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("чтение базовой спецификации: %w", err)
	}

	var parsed any
	err = yaml.Unmarshal(data, &parsed)
	if err != nil {
		return nil, fmt.Errorf("%s разбор базовой спецификации: %w", path, err)
	}

	root, er := parsed.(map[string]any)
	if !er {
		return nil, fmt.Errorf("%s базовая спецификация должен быть map[string]any", path)
	}

	comp := []string{"openapi", "info", "paths", "components"}

	for _, field := range comp {
		_, ok := root[field]
		if !ok {
			return nil, fmt.Errorf("%s в базовой спецификации нет обязательного поля %q", path, field)
		}
	}

	return &Document{Root: root, Path: path}, nil
}
