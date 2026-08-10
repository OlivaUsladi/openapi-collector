package merge

import (
	"fmt"
	"openapi-collector/internal/model"
)

var baseFields = []string{"openapi", "info", "servers", "externalDocs", "security"}

type Conflict struct {
	Kind   string
	Key    string
	First  model.Origin
	Second model.Origin
}

/*
Ожидаемая ошибка:
duplicate component components.schemas.Task:
  first definition: domain/task.go:3:5
  second definition: api/task.go:3:5
*/

func (c Conflict) String() string {
	return fmt.Sprintf("duplicate %s %s:\n  first definition: %s\n  second definition: %s",
		c.Kind, c.Key, formatOrigin(c.First), formatOrigin(c.Second))
}

func formatOrigin(o model.Origin) string {
	if o.Line == 0 {
		return o.File
	}
	return fmt.Sprintf("%s:%d:%d", o.File, o.Line, o.Column)
}
