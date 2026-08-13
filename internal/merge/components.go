package merge

import (
	"openapi-collector/internal/model"
	"strings"
)

var allowedSections = map[string]bool{
	"schemas": true, "responses": true, "parameters": true, "examples": true,
	"requestBodies": true, "headers": true, "securitySchemes": true,
	"links": true, "callbacks": true,
}

// проверка надопустимые симолы
func isValidComponentName(name string) bool {
	if name == "" {
		return false
	}

	for _, ch := range name {
		isLetter := (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
		isDigit := ch >= '0' && ch <= '9'
		isAllowed := ch == '.' || ch == '_' || ch == '-'

		if !isLetter && !isDigit && !isAllowed {
			return false
		}
	}
	return true
}

// объединяет components из нового фрагмента с основным документом
func (m *merger) mergeComponents(value any, origin model.Origin) {
	components, ok := value.(map[string]any)
	if !ok {
		m.errorf("%s:%d: фрагмент \"components\" должен быть map",
			origin.File, origin.Line)
		return
	}

	docComponents, ok := m.res.Doc["components"].(map[string]any)
	if !ok {
		docComponents = map[string]any{}
		m.res.Doc["components"] = docComponents
	}

	for sectionName, sectionData := range components {
		if !allowedSections[sectionName] && !strings.HasPrefix(sectionName, "x-") {
			m.errorf("%s:%d: неизвестный компонент секции %q", origin.File, origin.Line, sectionName)
			continue
		}

		sectionItems, ok := sectionData.(map[string]any)
		if !ok {
			m.errorf("%s:%d: компонент секции %q должен быть map",
				origin.File, origin.Line, sectionName)
			continue
		}

		docSection, ok := docComponents[sectionName].(map[string]any)
		if !ok {
			docSection = map[string]any{}
			docComponents[sectionName] = docSection
		}

		for componentName, componentDef := range sectionItems {
			if !isValidComponentName(componentName) {
				m.errorf("%s:%d:неправильное имя компонента %q в секции %q",
					origin.File, origin.Line, componentName, sectionName)
				continue
			}

			key := componentKey(sectionName, componentName)
			first, exists := m.res.Owners[key]
			if exists {
				m.conflict("component", "components."+sectionName+"."+componentName, first, origin)
				continue
			}
			docSection[componentName] = componentDef
			m.res.Owners[key] = origin

		}
	}
}

// объединяет теги из нового фрагмента с основным документом
func (m *merger) mergeTags(value any, origin model.Origin) {
	tags, ok := value.([]any)
	if !ok {
		m.errorf("%s:%d: поле \"tags\" должен быть массивом", origin.File, origin.Line)
		return
	}

	docTags, _ := m.res.Doc["tags"].([]any)

	for _, tagData := range tags {
		tag, ok := tagData.(map[string]any)
		if !ok {
			m.errorf("%s:%d: тег должен быть map", origin.File, origin.Line)
			continue
		}

		tagName, ok := tag["name"].(string)
		if !ok || tagName == "" {
			m.errorf("%s:%d: тег должен иметь \"name\"", origin.File, origin.Line)
			continue
		}

		key := tagKey(tagName)
		first, exists := m.res.Owners[key]
		if exists {
			existingTag := findTag(docTags, tagName)

			if !deepEqual(any(existingTag), any(tag)) {
				m.conflict("tag", `"`+tagName+`"`, first, origin)
			}
			continue
		}

		docTags = append(docTags, tag)
		m.res.Owners[key] = origin
	}

	if len(docTags) > 0 {
		m.res.Doc["tags"] = docTags
	}
}

// ищет тег по имени в массиве тегов
func findTag(tags []any, name string) map[string]any {
	for _, tagData := range tags {
		tag, ok := tagData.(map[string]any)
		if ok {
			tagName, _ := tag["name"].(string)
			if tagName == name {
				return tag
			}
		}
	}
	return nil
}
