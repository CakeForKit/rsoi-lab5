package model

import "github.com/google/uuid"

type Ticket struct {
	BaseEntity
	TicketUID    uuid.UUID `json:"ticketUid" gorm:"type:uuid;uniqueIndex;not null"`
	Username     string    `json:"-" gorm:"not null;index"`
	FlightNumber string    `json:"flightNumber" gorm:"not null"`
	Price        int       `json:"price" gorm:"not null"`
	Status       string    `json:"status" gorm:"not null"`
}

type TicketPurchaseRequest struct {
	FlightNumber    string `json:"flightNumber"`
	Price           int    `json:"price"`
	PaidFromBalance bool   `json:"paidFromBalance"`
}

type TicketCreateRequest struct {
	FlightNumber string `json:"flightNumber"`
	Price        int    `json:"price"`
}

type TicketResponse struct {
	TicketUID    uuid.UUID `json:"ticketUid"`
	FlightNumber string    `json:"flightNumber"`
	FromAirport  string    `json:"fromAirport"`
	ToAirport    string    `json:"toAirport"`
	Date         string    `json:"date"`
	Price        int       `json:"price"`
	Status       string    `json:"status"`
}

type TicketPurchaseResponse struct {
	TicketResponse
	PaidByMoney   int                `json:"paidByMoney"`
	PaidByBonuses int                `json:"paidByBonuses"`
	Privilege     PrivilegeShortInfo `json:"privilege"`
}
