package models

type Animal struct {
	Name        string `json:"name"`
	Type        int    `json:"type"`
	Description string `json:"description"`
}

// AnimalWithID - one record processed into json parseable object.
type AnimalWithID struct {
	ID     int    `json:"id"`
	Animal Animal `json:"data"`
}
