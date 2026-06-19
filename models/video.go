package models

import (
	"time"

	"github.com/google/uuid"
)

type Video struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Title     string    `gorm:"not null" json:"title"`
	FilePath  string    `gorm:"not null" json:"file_path"`
	Duration  float64   `gorm:"default:0" json:"duration"`
	FileSize  int64     `json:"file_size"`
	Status    string    `gorm:"default:'uploaded'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VideoClip struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	VideoID     uuid.UUID `gorm:"type:uuid;not null;index" json:"video_id"`
	Title       string    `json:"title"`
	FilePath    string    `json:"file_path"`
	StartTime   float64   `json:"start_time"`
	EndTime     float64   `json:"end_time"`
	Duration    float64   `json:"duration"`
	Confidence  float64   `json:"confidence"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}

type ClipJob struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	VideoID       uuid.UUID `gorm:"type:uuid;not null;index" json:"video_id"`
	Status        string    `gorm:"default:'pending'" json:"status"`
	Progress      int       `gorm:"default:0" json:"progress"`
	ClipsCreated  int       `gorm:"default:0" json:"clips_created"`
	ErrorMessage  string    `json:"error_message"`
	SceneMinGap   float64   `gorm:"default:2.0" json:"scene_min_gap"`
	MinClipLength float64   `gorm:"default:5.0" json:"min_clip_length"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (Video) TableName() string {
	return "videos"
}

func (VideoClip) TableName() string {
	return "video_clips"
}

func (ClipJob) TableName() string {
	return "clip_jobs"
}
