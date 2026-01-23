package entity

type User struct {
	ID                  uint   `gorm:"primary_key;auto_increment"`
	Issuer              string `gorm:"not null;index:idx_issuer_sub, unique"`
	Sub                 string `gorm:"not null; index:idx_issuer_sub, unique"`
	Name                string
	Preferred_username  string
	Given_name          string
	Family_name         string
	Email               string
	OnboardingCompleted bool
}

func NewUser() *User {
	return &User{
		OnboardingCompleted: false,
	}
}
