package models

import (
	"time"

	"github.com/google/uuid"
)

type PaymentProof struct {
	ID                 uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TransactionID      uuid.UUID           `gorm:"type:uuid;not null;index" json:"transaction_id"`
	UserID             uuid.UUID           `gorm:"type:uuid;not null;index" json:"user_id"`
	AccountName        string              `gorm:"not null" json:"account_name"`
	SourceBank         string              `gorm:"not null" json:"source_bank"`
	SourceAccountNumber string             `json:"source_account_number"`
	Status             string              `gorm:"default:'pending';not null" json:"status"`
	AdminNotes         string              `gorm:"type:text" json:"admin_notes"`
	RefundProofURL     string              `json:"refund_proof_url"`
	RefundAmount       float64             `json:"refund_amount"`
	Photos             []PaymentProofPhoto `gorm:"foreignKey:PaymentProofID" json:"photos"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

func (PaymentProof) TableName() string {
	return "payment_proofs"
}

type PaymentProofPhoto struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PaymentProofID  uuid.UUID `gorm:"type:uuid;not null;index" json:"payment_proof_id"`
	PhotoURL        string    `gorm:"not null" json:"photo_url"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

func (PaymentProofPhoto) TableName() string {
	return "payment_proof_photos"
}
