package model

// YAML-фрагмент из комментария @openapi
type Fragment struct {
	Origin   Origin   `json:"origin"`
	Sections []string `json:"sections"`
	Raw      string   `json:"-"`
}
