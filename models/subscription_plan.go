package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type SubscriptionPlan struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name             string    `gorm:"not null" json:"name"`
	Description      string    `json:"description"`
	Benefits         string    `gorm:"type:text" json:"-"`
	BenefitsList     []string  `gorm:"-" json:"benefits"`
	Price            float64   `gorm:"not null" json:"price"`
	DiscountPercent  float64   `gorm:"default:0" json:"discount_percent"`
	FinalPrice       float64   `gorm:"not null" json:"final_price"`
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

func (p *SubscriptionPlan) PrepareResponse() {
	p.BenefitsList = strings.Split(p.Benefits, ",")
	for i := range p.BenefitsList {
		p.BenefitsList[i] = strings.TrimSpace(p.BenefitsList[i])
	}
	origPrice := p.Price
	if p.DiscountPercent > 0 {
		if p.FinalPrice > 0 {
			origPrice = p.FinalPrice
		} else {
			origPrice = p.Price / (1 - p.DiscountPercent/100)
		}
	}
	p.FinalPrice = p.Price
	p.Price = origPrice
}
