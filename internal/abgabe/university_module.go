package abgabe

import "gorm.io/gorm"

type UniversityModule struct {
	gorm.Model
	Name string `json:"name"`
	Code string `json:"code"`
}
