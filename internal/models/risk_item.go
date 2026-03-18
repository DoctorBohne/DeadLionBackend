package models

import "time"

type RiskItem struct {
	ID        uint
	Type      string
	Title     string
	Deadline  time.Time
	Priority  int
	RiskScore float64

	//optional fields
	Room *string `json:"room,omitempty"`
}
