package refs

import (
	"fmt"
	"openapi-collector/internal/model"
	"sort"
	"strings"
)

type Ref struct {
	Target  string
	DocPath string
	Origin  model.Origin
}

/*
api/list_tasks.go:15: unresolved reference: #/components/schemas/UnknownTask
*/

// проверяет все ссылки $ref в документе после завершения merge
func Check(doc map[string]any, owners map[string]model.Origin) (errs []error, warnings []string) {
	collected := collect(doc, owners)

	used := map[string]bool{}
	for _, ref := range collected {
		if !strings.HasPrefix(ref.Target, "#/") {
			errs = append(errs, fmt.Errorf("%s: external reference is not supported: %s",
				formatOrigin(ref.Origin), ref.Target))
			continue
		}
		section, name, ok := splitLocalRef(ref.Target)
		if !ok {
			errs = append(errs, fmt.Errorf("%s: unresolved reference: %s",
				formatOrigin(ref.Origin), ref.Target))
			continue
		}
		if !componentExists(doc, section, name) {
			errs = append(errs, fmt.Errorf("%s: unresolved reference: %s",
				formatOrigin(ref.Origin), ref.Target))
			continue
		}
		used["comp:"+section+"."+name] = true
	}

	components, _ := doc["components"].(map[string]any)
	for section, rawNames := range components {
		names, ok := rawNames.(map[string]any)
		if !ok {
			continue
		}
		for name := range names {
			if section == "securitySchemes" {
				continue
			}
			if !used["comp:"+section+"."+name] {
				warnings = append(warnings,
					fmt.Sprintf("unused component components.%s.%s", section, name))
			}
		}
	}

	sort.Slice(errs, func(i, j int) bool {
		return errs[i].Error() < errs[j].Error()
	})
	sort.Strings(warnings)
	return errs, warnings
}

// рекурсивно обходит документ и собирает все $ref
func collect(doc map[string]any, owners map[string]model.Origin) []Ref {
	var out []Ref
	walk(doc, "", owners, model.Origin{}, &out)
	sort.Slice(out, func(i, j int) bool {
		if out[i].DocPath != out[j].DocPath {
			return out[i].DocPath < out[j].DocPath
		}
		return out[i].Target < out[j].Target
	})
	return out
}

// рекурсивный обход значения
// обходит структуру данных и находит все $ref ссылки
func walk(value any, currentPath string, owners map[string]model.Origin, currentOwner model.Origin, out *[]Ref) {
	dictionary, ok := value.(map[string]any)
	if ok {
		for key, item := range dictionary {
			nextPath := key
			if currentPath != "" {
				nextPath = currentPath + "." + key
			}

			owner, found := ownerFor(nextPath, owners)
			if found {
				currentOwner = owner
			}

			if key == "$ref" {
				target, ok := item.(string)
				if ok {
					ref := Ref{
						Target:  target,
						DocPath: currentPath,
						Origin:  currentOwner,
					}
					*out = append(*out, ref)
					continue
				}
			}

			walk(item, nextPath, owners, currentOwner, out)
		}
		return
	}

	array, ok := value.([]any)
	if ok {
		for index, item := range array {
			itemPath := fmt.Sprintf("%s[%d]", currentPath, index)

			walk(item, itemPath, owners, currentOwner, out)
		}
		return
	}
}

// сопоставляет путь внутри документа с ключом карты владельцев
func ownerFor(docPath string, owners map[string]model.Origin) (model.Origin, bool) {
	parts := strings.Split(docPath, ".")
	if len(parts) == 3 && parts[0] == "paths" {
		o, ok := owners["op:"+parts[2]+" "+parts[1]]
		if ok {
			return o, true
		}
	}
	if len(parts) == 3 && parts[0] == "components" {
		o, ok := owners["comp:"+parts[1]+"."+parts[2]]
		if ok {
			return o, true
		}
	}
	return model.Origin{}, false
}

// разбирает "#/components/schemas/Task" на раздел и имя
func splitLocalRef(target string) (section, name string, ok bool) {
	parts := strings.Split(strings.TrimPrefix(target, "#/"), "/")
	if len(parts) != 3 || parts[0] != "components" {
		return "", "", false
	}
	return parts[1], parts[2], true
}

// проверяет наличие компонента в документе
func componentExists(doc map[string]any, section, name string) bool {
	components, ok := doc["components"].(map[string]any)
	if !ok {
		return false
	}
	names, ok := components[section].(map[string]any)
	if !ok {
		return false
	}
	_, ok = names[name]
	return ok
}

// печатает источник фрагмента для сообщения об ошибке
func formatOrigin(o model.Origin) string {
	if o.File == "" {
		return "document"
	}
	if o.Line == 0 {
		return o.File
	}
	return fmt.Sprintf("%s:%d", o.File, o.Line)
}
