package model

// место, из которого извлечён фрагмент спецификации
type Origin struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}
