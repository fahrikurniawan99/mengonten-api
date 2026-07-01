package models

import (
	"time"

	"github.com/google/uuid"
)

type SceneJob struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	YouTubeURL string    `gorm:"column:youtube_url;not null" json:"youtube_url"`
	Title      string    `json:"title"`
	Duration   float64   `json:"duration"`
	Thumbnail  string    `json:"thumbnail"`
	Tags       []byte    `gorm:"type:jsonb" json:"tags"`
	Categories []byte    `gorm:"type:jsonb" json:"categories"`
	IsLive     bool      `json:"is_live"`
	Status     string    `gorm:"default:'pending';not null" json:"status"`
	Progress   int       `gorm:"default:0" json:"progress"`
	Transcript string    `gorm:"type:text" json:"transcript,omitempty"`
	Chapters   []byte    `gorm:"type:jsonb" json:"chapters"`
	Error      string    `json:"error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (SceneJob) TableName() string {
	return "scene_jobs"
}

type Chapter struct {
	Index       int      `json:"index"`
	StartTime   float64  `json:"start_time"`
	EndTime     float64  `json:"end_time"`
	Duration    float64  `json:"duration"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}
