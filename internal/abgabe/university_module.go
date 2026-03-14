package abgabe

import "gorm.io/gorm"

// This Struct contains information regarding university modules
// Abgaben can be linked to this module (one to many)
type UniversityModule struct {
	gorm.Model
	Name string `json:"name"`
	Code string `json:"code"`
}
