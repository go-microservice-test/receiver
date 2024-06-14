package models

type Animal struct {
	Name        string `json:"name"`
	Type        int    `json:"type"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive"`
}
