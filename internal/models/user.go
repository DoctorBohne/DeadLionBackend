package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Sub                 string `gorm:"not null;index:idx_users_iss_sub,unique"`
	Issuer              string `gorm:"not null;index:idx_users_iss_sub,unique"`
	EmailVerified       bool
	Name                string
	PreferredUsername   string
	GivenName           string
	FamilyName          string
	Email               string
	onboardingCompleted bool `gorm:"not null;default:false"`
}
