package extract

import (
	"fmt"
	"openapi-collector/internal/model"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

func buildFragment(raw rawFragment) (model.Fragment, error) {
	text := strings.Join(raw.lines, "\n")

	var parsed any
	err := yaml.Unmarshal([]byte(text), &parsed)
	if err != nil {
		return model.Fragment{}, fmt.Errorf("%s:%d:%d: парсинг @openapi %w",
			raw.origin.File, raw.origin.Line, raw.origin.Column, err)
	}
	if parsed == nil {
		return model.Fragment{}, fmt.Errorf("%s:%d:%d: пустой @openapiнет YAML после маркера",
			raw.origin.File, raw.origin.Line, raw.origin.Column)
	}

	doc, ok := parsed.(map[string]any)
	if !ok {
		return model.Fragment{}, fmt.Errorf("%s:%d:%d: @openapi фрагемент не map",
			raw.origin.File, raw.origin.Line, raw.origin.Column)
	}

	sections := make([]string, 0, len(doc))
	for key := range doc {
		sections = append(sections, key)
	}
	sort.Strings(sections)

	return model.Fragment{
		Origin:   raw.origin,
		Sections: sections,
		Raw:      text,
		Doc:      doc,
	}, nil
}
