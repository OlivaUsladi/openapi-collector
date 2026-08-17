package output

import (
	"sort"
	"strings"
)

var topLevelOrder = []string{
	"openapi", "info", "servers", "externalDocs", "tags", "security", "paths", "components",
}

var methodOrder = []string{"get", "post", "put", "patch", "delete", "options", "head", "trace"}

var pathItemOrder = append([]string{"summary", "description", "servers", "parameters"}, methodOrder...)

var componentsOrder = []string{
	"schemas", "responses", "parameters", "examples", "requestBodies",
	"headers", "securitySchemes", "links", "callbacks",
}

func keyLess(order []string) func(a, b string) bool {
	rank := make(map[string]int, len(order))
	for i, key := range order {
		rank[key] = i
	}

	return func(a, b string) bool {
		aInOrder := false
		bInOrder := false
		aRank := 0
		bRank := 0

		rankA, ok := rank[a]
		if ok {
			aInOrder = true
			aRank = rankA
		}

		rankB, ok := rank[b]
		if ok {
			bInOrder = true
			bRank = rankB
		}

		aIsExt := strings.HasPrefix(a, "x-")
		bIsExt := strings.HasPrefix(b, "x-")

		if aInOrder && bInOrder {
			return aRank < bRank
		}

		if aInOrder {
			return true
		}

		if bInOrder {
			return false
		}

		if aIsExt && !bIsExt {
			return false
		}

		if !aIsExt && bIsExt {
			return true
		}

		return a < b
	}
}

// возвращает ключи отображения в заданном порядке
func sortedKeys(m map[string]any, less func(a, b string) bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return less(keys[i], keys[j])
	})
	return keys
}

// выбирает правило порядка ключей по месту в документе
func orderFor(docPath string) func(a, b string) bool {
	pathParts := strings.Split(docPath, ".")

	if docPath == "" {
		return keyLess(topLevelOrder)
	}

	if docPath == "components" {
		return keyLess(componentsOrder)
	}

	if len(pathParts) == 2 && pathParts[0] == "paths" {
		return keyLess(pathItemOrder)
	}

	return keyLess(nil)
}
