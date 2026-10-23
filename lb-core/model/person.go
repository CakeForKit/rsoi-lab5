package model

type Person struct {
	BaseEntity
	Name    string `json:"name" gorm:"not null"`
	Age     int    `json:"age"`
	Address string `json:"address"`
	Work    string `json:"work"`
}

type PersonUpdate struct {
	Name    *string `json:"name"`
	Age     *int    `json:"age"`
	Address *string `json:"address"`
	Work    *string `json:"work"`
}
