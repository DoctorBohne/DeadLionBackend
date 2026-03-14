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
	ModuleID uint    `json:"module,omitempty"`
	Room     *string `json:"room,omitempty"`
}
