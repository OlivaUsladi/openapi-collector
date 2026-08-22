package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"openapi-collector/internal/merge"
	"openapi-collector/internal/model"
	"sort"
	"strings"
)

const Version = 1

type Params struct {
	Base    string   `json:"base"`
	Source  string   `json:"source"`
	Format  string   `json:"format"`
	Exclude []string `json:"exclude,omitempty"`
	Strict  bool     `json:"strict"`
}

type Totals struct {
	GoFiles    int `json:"go_files"`
	Fragments  int `json:"fragments"`
	Operations int `json:"operations"`
	Components int `json:"components"`
	Tags       int `json:"tags"`
	Conflicts  int `json:"conflicts"`
	Warnings   int `json:"warnings"`
}

type Fragment struct {
	Origin   model.Origin `json:"origin"`
	Sections []string     `json:"sections"`
}

type Operation struct {
	Method string       `json:"method"`
	Path   string       `json:"path"`
	Origin model.Origin `json:"origin"`
}

type Component struct {
	Section string       `json:"section"`
	Name    string       `json:"name"`
	Origin  model.Origin `json:"origin"`
}

type Conflict struct {
	Kind   string       `json:"kind"`
	Key    string       `json:"key"`
	First  model.Origin `json:"first"`
	Second model.Origin `json:"second"`
}

type Warning struct {
	Kind   string `json:"kind"`
	Target string `json:"target"`
}

type Report struct {
	ReportVersion int         `json:"report_version"`
	Params        Params      `json:"params"`
	Totals        Totals      `json:"totals"`
	Fragments     []Fragment  `json:"fragments"`
	Operations    []Operation `json:"operations"`
	Components    []Component `json:"components"`
	Tags          []string    `json:"tags"`
	Conflicts     []Conflict  `json:"conflicts"`
	Warnings      []Warning   `json:"warnings"`
	Errors        []string    `json:"errors"`
}

// собирает отчёт из результатов конвейера
func Build(params Params, goFiles int, fragments []model.Fragment,
	res *merge.Result, warnings []string, errs []error) *Report {

	r := &Report{
		ReportVersion: Version,
		Params:        params,
		Fragments:     []Fragment{},
		Operations:    []Operation{},
		Components:    []Component{},
		Tags:          []string{},
		Conflicts:     []Conflict{},
		Warnings:      []Warning{},
		Errors:        []string{},
	}

	for _, frag := range fragments {
		r.Fragments = append(r.Fragments, Fragment{Origin: frag.Origin, Sections: frag.Sections})
	}
	sort.Slice(r.Fragments, func(i, j int) bool {
		a, b := r.Fragments[i].Origin, r.Fragments[j].Origin
		if a.File != b.File {
			return a.File < b.File
		}
		return a.Line < b.Line
	})

	// хранит ключи "op:get /tasks", "comp:schemas.Task", "tag:tasks".
	for key, origin := range res.Owners {
		if strings.HasPrefix(key, "op:") {
			rest := strings.TrimPrefix(key, "op:")

			method, path, _ := strings.Cut(rest, " ")

			r.Operations = append(r.Operations, Operation{
				Method: method,
				Path:   path,
				Origin: origin,
			})
			continue
		}

		if strings.HasPrefix(key, "comp:") {
			rest := strings.TrimPrefix(key, "comp:")

			section, name, _ := strings.Cut(rest, ".")

			r.Components = append(r.Components, Component{
				Section: section,
				Name:    name,
				Origin:  origin,
			})
			continue
		}

		if strings.HasPrefix(key, "tag:") {
			tagName := strings.TrimPrefix(key, "tag:")
			r.Tags = append(r.Tags, tagName)
			continue
		}

	}
	sort.Slice(r.Operations, func(i, j int) bool {
		if r.Operations[i].Path != r.Operations[j].Path {
			return r.Operations[i].Path < r.Operations[j].Path
		}
		return r.Operations[i].Method < r.Operations[j].Method
	})
	sort.Slice(r.Components, func(i, j int) bool {
		if r.Components[i].Section != r.Components[j].Section {
			return r.Components[i].Section < r.Components[j].Section
		}
		return r.Components[i].Name < r.Components[j].Name
	})
	sort.Strings(r.Tags)

	for _, c := range res.Conflicts {
		r.Conflicts = append(r.Conflicts, Conflict{
			Kind: c.Kind, Key: c.Key, First: c.First, Second: c.Second,
		})
	}

	for _, w := range warnings {
		kind := "warning"
		if strings.HasPrefix(w, "unused component ") {
			kind = "unused_component"
		}
		r.Warnings = append(r.Warnings, Warning{
			Kind:   kind,
			Target: strings.TrimPrefix(w, "unused component "),
		})
	}

	for _, e := range errs {
		r.Errors = append(r.Errors, e.Error())
	}
	sort.Strings(r.Errors)

	r.Totals = Totals{
		GoFiles:    goFiles,
		Fragments:  len(r.Fragments),
		Operations: len(r.Operations),
		Components: len(r.Components),
		Tags:       len(r.Tags),
		Conflicts:  len(r.Conflicts),
		Warnings:   len(r.Warnings),
	}
	return r
}

// сериализует отчёт в JSON с отступом 2 пробела
func MarshalIndent(r *Report) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	err := enc.Encode(r)
	if err != nil {
		return nil, fmt.Errorf("сериализация отчёта: %w", err)
	}
	return buf.Bytes(), nil
}

// читает JSON-отчёт из данных
func Load(data []byte) (*Report, error) {
	var r Report
	err := json.Unmarshal(data, &r)
	if err != nil {
		return nil, fmt.Errorf("разбор JSON-отчёта: %w", err)
	}
	if r.ReportVersion != Version {
		return nil, fmt.Errorf("неподдерживаемая версия отчёта: %d", r.ReportVersion)
	}
	return &r, nil
}
