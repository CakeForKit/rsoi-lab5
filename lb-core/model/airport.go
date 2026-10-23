package model

type Airport struct {
	BaseEntity
	Name    string `json:"name" gorm:"not null"`
	City    string `json:"city" gorm:"not null"`
	Country string `json:"country" gorm:"not null"`
}
