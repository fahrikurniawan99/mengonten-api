package models

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	UserSubscriptionID *uuid.UUID `gorm:"type:uuid;index" json:"user_subscription_id"`
	ReferenceID       string     `gorm:"uniqueIndex" json:"reference_id"`
	Status            string     `gorm:"default:'pending';not null" json:"status"`
	Amount            float64    `gorm:"not null" json:"amount"`
	UniqueCode        int        `gorm:"not null" json:"unique_code"`
	TotalAmount       float64    `gorm:"not null" json:"total_amount"`
	BankName          string     `gorm:"not null" json:"bank_name"`
	BankAccountNumber string     `gorm:"not null" json:"bank_account_number"`
	BankAccountName   string     `gorm:"not null" json:"bank_account_name"`
	PaidAt            *time.Time `json:"paid_at"`
	ExpiredAt         *time.Time `json:"expired_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (Transaction) TableName() string {
	return "transactions"
}

func (t *Transaction) PrepareResponse() {
}

type UserSubscription struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TransactionID uuid.UUID `gorm:"type:uuid;not null;index" json:"transaction_id"`
	PlanName     string     `gorm:"not null" json:"plan_name"`
	PlanType     string     `json:"plan_type"`
	PlanBenefits string     `gorm:"type:text" json:"-"`
	BenefitsList []string   `gorm:"-" json:"plan_benefits"`
	PlanPrice      float64    `json:"plan_price"`
	PlanDuration   int        `json:"plan_duration"`
	StorageUsedBytes int64    `gorm:"default:0" json:"storage_used_bytes"`
	Rules          string     `gorm:"type:text" json:"-"`
	RulesMap       map[string]string `gorm:"-" json:"rules"`
	Status         string     `gorm:"default:'active';not null" json:"status"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      time.Time  `json:"end_date"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (UserSubscription) TableName() string {
	return "user_subscriptions"
}

func (s *UserSubscription) PrepareResponse() {
	s.BenefitsList = parseStringToList(s.PlanBenefits)
	if s.Rules != "" {
		s.RulesMap = make(map[string]string)
		json.Unmarshal([]byte(s.Rules), &s.RulesMap)
	}
	if s.RulesMap == nil {
		s.RulesMap = make(map[string]string)
	}
}

func (s *UserSubscription) IsActive() bool {
	return s.Status == "active" && time.Now().Before(s.EndDate)
}

type TransactionPreview struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ReferenceID          string    `gorm:"uniqueIndex;not null" json:"reference_id"`
	UserID               uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	SubscriptionPlanID   uuid.UUID `gorm:"type:uuid;not null" json:"subscription_plan_id"`
	BankAccountID        uuid.UUID `gorm:"type:uuid;not null" json:"bank_account_id"`
	Amount               float64   `gorm:"not null" json:"amount"`
	UniqueCode           int       `gorm:"not null" json:"unique_code"`
	TotalAmount          float64   `gorm:"not null" json:"total_amount"`
	BankName             string    `gorm:"not null" json:"bank_name"`
	BankAccountNumber    string    `gorm:"not null" json:"bank_account_number"`
	BankAccountName      string    `gorm:"not null" json:"bank_account_name"`
	SubscriptionName     string    `gorm:"not null" json:"subscription_name"`
	SubscriptionType     string    `json:"subscription_type"`
	SubscriptionDuration int       `json:"subscription_duration"`
	SubscriptionBenefits string    `gorm:"type:text" json:"-"`
	BenefitsList         []string  `gorm:"-" json:"subscription_benefits"`
	ExpiresAt            time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt            time.Time `json:"created_at"`
}

func (TransactionPreview) TableName() string {
	return "transaction_previews"
}

func (p *TransactionPreview) PrepareResponse() {
	p.BenefitsList = parseStringToList(p.SubscriptionBenefits)
}

func (p *TransactionPreview) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

func parseStringToList(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
