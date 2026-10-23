package model

type BaseEntity struct {
	ID uint `json:"id,omitempty" gorm:"primaryKey;autoIncrement"`
}
