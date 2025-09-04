package main

import "time"

type Account struct {
	ID             uint   `json:"account_id" gorm:"primaryKey"`
	DocumentNumber string `json:"document_number" gorm:"uniqueIndex;not null"`
}

type OperationType struct {
	ID          int    `json:"operation_type_id" gorm:"primaryKey"`
	Description string `json:"description"`
}

type Transaction struct {
	ID              uint      `json:"transaction_id" gorm:"primaryKey"`
	AccountID       uint      `json:"account_id" gorm:"not null"`
	OperationTypeID int       `json:"operation_type_id" gorm:"not null"`
	Amount          float64   `json:"amount" gorm:"not null"`
	EventDate       time.Time `json:"event_date" gorm:"autoCreateTime"`
}
