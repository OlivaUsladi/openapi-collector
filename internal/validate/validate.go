package validate

import (
	"fmt"
	"openapi-collector/internal/model"
	"sort"
	"strings"
)

var httpMethods = map[string]bool{
	"get": true, "post": true, "put": true, "patch": true,
	"delete": true, "options": true, "head": true, "trace": true,
}

// валидирует собранный документ как OpenAPI 3.0
func Check(doc map[string]any, owners map[string]model.Origin) []error {
	var errs []error

	errs = append(errs, checkInfo(doc)...)
	errs = append(errs, checkOperations(doc, owners)...)
	errs = append(errs, checkOperationIDs(doc, owners)...)
	errs = append(errs, checkTemplatePaths(doc, owners)...)
	errs = append(errs, checkSecurity(doc)...)

	sort.Slice(errs, func(i, j int) bool {
		return errs[i].Error() < errs[j].Error()
	})
	return errs
}

// проверяет openapi, info.title и info.version
func checkInfo(doc map[string]any) []error {
	var errs []error

	version, _ := doc["openapi"].(string)
	if !strings.HasPrefix(version, "3.0") {
		errs = append(errs, fmt.Errorf(`field "openapi" must be a 3.0.x version, got %q`, version))
	}

	info, ok := doc["info"].(map[string]any)
	if !ok {
		return append(errs, fmt.Errorf(`field "info" must be a mapping`))
	}

	title, _ := info["title"].(string)
	if title == "" {
		errs = append(errs, fmt.Errorf(`field "info.title" is required`))
	}

	v, _ := info["version"].(string)
	if v == "" {
		errs = append(errs, fmt.Errorf(`field "info.version" is required`))
	}

	return errs
}

// вызывает fn для каждой операции документа
func operations(doc map[string]any, fn func(path, method string, op map[string]any)) {
	paths, _ := doc["paths"].(map[string]any)
	for path, rawItem := range paths {
		item, ok := rawItem.(map[string]any)
		if !ok {
			continue
		}
		for field, rawOp := range item {
			if !httpMethods[field] {
				continue
			}
			op, ok := rawOp.(map[string]any)
			if !ok {
				continue
			}
			fn(path, field, op)
		}
	}
}

// возвращает источник операции для сообщения об ошибке
func origin(owners map[string]model.Origin, method, path string) string {
	o, ok := owners["op:"+method+" "+path]
	if !ok {
		return ""
	}
	if o.Line == 0 {
		return " (" + o.File + ")"
	}
	return fmt.Sprintf(" (%s:%d)", o.File, o.Line)
}

// проверяет обязательное поле responses каждой операции
func checkOperations(doc map[string]any, owners map[string]model.Origin) []error {
	var errs []error
	operations(doc, func(path, method string, op map[string]any) {
		responses, ok := op["responses"].(map[string]any)
		if !ok || len(responses) == 0 {
			errs = append(errs, fmt.Errorf("operation %s %s has no responses%s",
				strings.ToUpper(method), path, origin(owners, method, path)))
		}
	})
	return errs
}

// проверяет уникальность operationId в документе
func checkOperationIDs(doc map[string]any, owners map[string]model.Origin) []error {
	type usage struct{ method, path string }
	seen := map[string]usage{}
	var errs []error

	operations(doc, func(path, method string, op map[string]any) {
		id, _ := op["operationId"].(string)
		if id == "" {
			return
		}

		first, ok := seen[id]
		if ok {
			errs = append(errs, fmt.Errorf(
				"duplicate operationId %q:\n  first definition: %s %s%s\n  second definition: %s %s%s",
				id,
				strings.ToUpper(first.method), first.path, origin(owners, first.method, first.path),
				strings.ToUpper(method), path, origin(owners, method, path)))
			return
		}
		seen[id] = usage{method, path}
	})

	return errs
}

// находит все шаблонные параметры в пути
func findTemplateParams(path string) []string {
	var params []string
	isInBrace := false
	currentParam := ""

	for i := 0; i < len(path); i++ {
		ch := path[i]

		if ch == '{' {
			isInBrace = true
			currentParam = ""
		} else if ch == '}' {
			isInBrace = false
			if currentParam != "" {
				params = append(params, currentParam)
			}
		} else if isInBrace {
			currentParam += string(ch)
		}
	}

	return params
}

// возвращает форму пути без имен параметров
func getPathShape(path string) string {
	result := ""
	isInBrace := false

	for i := 0; i < len(path); i++ {
		ch := path[i]

		if ch == '{' {
			isInBrace = true
			result += "{}"
		} else if ch == '}' {
			isInBrace = false
		} else if !isInBrace {
			result += string(ch)
		}
	}

	return result
}

// проверяет шаблонные параметры путей
func checkTemplatePaths(doc map[string]any, owners map[string]model.Origin) []error {
	var errs []error
	paths, _ := doc["paths"].(map[string]any)

	var sortedPaths []string
	for path := range paths {
		sortedPaths = append(sortedPaths, path)
	}
	sort.Strings(sortedPaths)

	shapes := map[string]string{}
	for _, path := range sortedPaths {
		shape := getPathShape(path)
		if shape == path {
			continue
		}

		first, ok := shapes[shape]
		if ok {
			errs = append(errs, fmt.Errorf(
				"equivalent paths %s and %s differ only in template names", first, path))
			continue
		}
		shapes[shape] = path
	}

	for _, path := range sortedPaths {
		item, ok := paths[path].(map[string]any)
		if !ok {
			continue
		}

		pathParams := declaredParams(doc, item["parameters"])

		for field, rawOp := range item {
			if !httpMethods[field] {
				continue
			}
			op, _ := rawOp.(map[string]any)

			declared := map[string]bool{}
			for name := range pathParams {
				declared[name] = true
			}
			for name := range declaredParams(doc, op["parameters"]) {
				declared[name] = true
			}

			for _, paramName := range findTemplateParams(path) {
				if !declared[paramName] {
					errs = append(errs, fmt.Errorf(
						"path parameter {%s} of %s %s is not described%s",
						paramName, strings.ToUpper(field), path, origin(owners, field, path)))
				}
			}
		}
	}

	return errs
}

// собирает имена path-параметров из массива parameters
func declaredParams(doc map[string]any, value any) map[string]bool {
	out := map[string]bool{}
	params, ok := value.([]any)
	if !ok {
		return out
	}

	for _, rawParam := range params {
		param, ok := rawParam.(map[string]any)
		if !ok {
			continue
		}

		ref, ok := param["$ref"].(string)
		if ok {
			param = resolveParameter(doc, ref)
			if param == nil {
				continue
			}
		}

		in, _ := param["in"].(string)
		if in != "path" {
			continue
		}

		name, _ := param["name"].(string)
		if name != "" {
			out[name] = true
		}
	}

	return out
}

// возвращает параметр по локальной ссылке
func resolveParameter(doc map[string]any, ref string) map[string]any {
	const prefix = "#/components/parameters/"
	if !strings.HasPrefix(ref, prefix) {
		return nil
	}

	components, _ := doc["components"].(map[string]any)
	params, _ := components["parameters"].(map[string]any)
	paramName := strings.TrimPrefix(ref, prefix)
	param, _ := params[paramName].(map[string]any)
	return param
}

// проверяет, что все security схемы существуют
func checkSecurity(doc map[string]any) []error {
	components, _ := doc["components"].(map[string]any)
	schemes, _ := components["securitySchemes"].(map[string]any)

	var errs []error

	checkSecurityRequirements(doc["security"], schemes, "global security", &errs)

	operations(doc, func(path, method string, op map[string]any) {
		where := strings.ToUpper(method) + " " + path
		checkSecurityRequirements(op["security"], schemes, where, &errs)
	})

	return errs
}

// проверяет список security требований
func checkSecurityRequirements(value any, schemes map[string]any, where string, errs *[]error) {
	requirements, ok := value.([]any)
	if !ok {
		return
	}

	for _, rawReq := range requirements {
		req, ok := rawReq.(map[string]any)
		if !ok {
			continue
		}

		for name := range req {
			_, exists := schemes[name]
			if !exists {
				*errs = append(*errs, fmt.Errorf(
					"security scheme %q used in %s is not defined in components.securitySchemes",
					name, where))
			}
		}
	}
}
