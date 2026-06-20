package models

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionPlan struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name             string    `gorm:"not null" json:"name"`
	Description      string    `json:"description"`
	Benefits         string    `gorm:"type:text" json:"benefits"`
	Price            float64   `gorm:"not null" json:"price"`
	DiscountPercent  float64   `gorm:"default:0" json:"discount_percent"`
	FinalPrice       float64   `gorm:"-" json:"final_price"`
	Type             string    `gorm:"not null" json:"type"`
	DurationDays     int       `gorm:"not null;default:30" json:"duration_days"`
	IsActive         bool      `gorm:"default:true" json:"is_active"`
	SortOrder        int       `gorm:"default:0" json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (SubscriptionPlan) TableName() string {
	return "subscription_plans"
}

func (p *SubscriptionPlan) CalculateFinalPrice() {
	if p.DiscountPercent > 0 {
		p.FinalPrice = p.Price - (p.Price * p.DiscountPercent / 100)
	} else {
		p.FinalPrice = p.Price
	}
}
