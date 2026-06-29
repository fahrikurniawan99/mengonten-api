package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Transaction struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	SubscriptionPlanID uuid.UUID `gorm:"type:uuid;not null" json:"subscription_plan_id"`
	ReferenceID       string     `gorm:"uniqueIndex;not null" json:"reference_id"`
	ProductName       string     `json:"product_name"`
	PaymentTotal      float64    `gorm:"not null" json:"payment_total"`
	Status            string     `gorm:"default:'pending';not null" json:"status"`
	PaymentMethod     string     `json:"payment_method"`
	PaymentNumber     string     `json:"payment_number"`
	ExpiredAt         *time.Time `json:"expired_at"`
	OrderID           *uuid.UUID `gorm:"type:uuid;index" json:"order_id"`
	PaymentAt         *time.Time `json:"payment_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (Transaction) TableName() string {
	return "transactions"
}

type Order struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TransactionID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"transaction_id"`
	PlanID            uuid.UUID  `gorm:"type:uuid;not null" json:"plan_id"`
	ProductName       string     `gorm:"not null" json:"product_name"`
	ProductPrice      float64    `gorm:"not null" json:"product_price"`
	Status            string     `gorm:"default:'pending';not null" json:"status"`
	ExpiredAt         time.Time  `json:"expired_at"`
	LastReminderSentAt *time.Time `json:"-"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	Rules             []OrderRule `gorm:"foreignKey:OrderID" json:"-"`
}

func (Order) TableName() string {
	return "orders"
}

func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ExpiredAt.IsZero() {
		var plan SubscriptionPlan
		if err := tx.First(&plan, o.PlanID).Error; err == nil {
			o.ExpiredAt = time.Now().AddDate(0, 0, plan.DurationDays)
		}
	}
	return
}

type OrderRule struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"order_id"`
	SubscriptionRuleID *uuid.UUID `gorm:"type:uuid;index" json:"subscription_rule_id"`
	RuleKey            string     `gorm:"not null" json:"rule_key"`
	RuleValue          string     `gorm:"not null" json:"rule_value"`
	CreatedAt          time.Time  `json:"created_at"`
}

func (OrderRule) TableName() string {
	return "order_rules"
}
