package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	PrivilegeOperationFillInBalance = "FILL_IN_BALANCE"
	PrivilegeOperationDebit         = "DEBIT_THE_ACCOUNT"
)

type Privilege struct {
	BaseEntity
	Username string `json:"-" gorm:"uniqueIndex;not null"`
	Status   string `json:"status" gorm:"not null"`
	Balance  int    `json:"balance" gorm:"not null"`
}

type PrivilegeHistory struct {
	BaseEntity
	PrivilegeID   uint      `json:"-" gorm:"not null;index"`
	TicketUID     uuid.UUID `json:"ticketUid" gorm:"type:uuid;not null"`
	DateTime      time.Time `json:"date" gorm:"not null"`
	BalanceDiff   int       `json:"balanceDiff" gorm:"not null"`
	OperationType string    `json:"operationType" gorm:"not null"`
}

type PrivilegeShortInfo struct {
	Balance int    `json:"balance,omitempty"`
	Status  string `json:"status,omitempty"`
}

type PrivilegeInfoResponse struct {
	Balance int                `json:"balance"`
	Status  string             `json:"status"`
	History []PrivilegeHistory `json:"history"`
}

type PrivilegeOperationRequest struct {
	TicketUID     uuid.UUID `json:"ticketUid"`
	Amount        int       `json:"amount"`
	OperationType string    `json:"operationType"`
}

type UserInfoResponse struct {
	Tickets   []TicketResponse   `json:"tickets"`
	Privilege PrivilegeShortInfo `json:"privilege"`
}
