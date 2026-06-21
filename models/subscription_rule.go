package models

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionRule struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PlanID      uuid.UUID `gorm:"type:uuid;not null;index" json:"plan_id"`
	RuleKey     string    `gorm:"not null" json:"rule_key"`
	RuleValue   string    `gorm:"not null" json:"rule_value"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (SubscriptionRule) TableName() string {
	return "subscription_rules"
}
