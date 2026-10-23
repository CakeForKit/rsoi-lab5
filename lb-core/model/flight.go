package model

import "time"

type Flight struct {
	BaseEntity
	FlightNumber  string    `json:"flightNumber" gorm:"uniqueIndex;not null"`
	DateTime      time.Time `json:"-" gorm:"not null"`
	FromAirportID uint      `json:"-"`
	ToAirportID   uint      `json:"-"`
	FromAirport   Airport   `json:"-" gorm:"foreignKey:FromAirportID"`
	ToAirport     Airport   `json:"-" gorm:"foreignKey:ToAirportID"`
	Price         int       `json:"price" gorm:"not null"`
}

type FlightResponse struct {
	FlightNumber string `json:"flightNumber"`
	FromAirport  string `json:"fromAirport"`
	ToAirport    string `json:"toAirport"`
	Date         string `json:"date"`
	Price        int    `json:"price"`
}

type PaginationResponse struct {
	Page          int              `json:"page"`
	PageSize      int              `json:"pageSize"`
	TotalElements int64            `json:"totalElements"`
	Items         []FlightResponse `json:"items"`
}
