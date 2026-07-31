package model

type Fragment struct {
	Origin   Origin   `json:"origin"`
	Sections []string `json:"sections"`
	Raw      string   `json:"-"`
}
