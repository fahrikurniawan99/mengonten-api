package models

import (
	"time"

	"github.com/google/uuid"
)

type BankAccount struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BankName      string    `gorm:"not null" json:"bank_name"`
	AccountNumber string    `gorm:"not null" json:"account_number"`
	AccountName   string    `gorm:"not null" json:"account_name"`
	Icon          string    `json:"icon"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (BankAccount) TableName() string {
	return "bank_accounts"
}
