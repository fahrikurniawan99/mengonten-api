package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email             string     `gorm:"uniqueIndex;not null" json:"email"`
	Username          string     `gorm:"uniqueIndex;not null" json:"username"`
	Password          string     `gorm:"not null" json:"-"`
	Role              string     `gorm:"default:'user';not null" json:"role"`
	IsVerified        bool       `gorm:"default:false" json:"is_verified"`
	VerificationToken string     `gorm:"index" json:"-"`
	TokenExpiresAt    *time.Time `json:"-"`
	VerifiedAt        *time.Time `json:"verified_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
