package models

import (
	"time"

	"github.com/google/uuid"
)

type YouTubeVideo struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	YouTubeURL     string     `gorm:"not null" json:"youtube_url"`
	Title          string     `json:"title"`
	Genre          string     `json:"genre"`
	Status         string     `gorm:"default:'pending'" json:"status"`
	LocalFilePath  string     `json:"local_file_path"`
	Duration       float64    `json:"duration"`
	FileSize       int64      `json:"file_size"`
	ProcessedAt    *time.Time `json:"processed_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type VideoTranscript struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	VideoID   uuid.UUID `gorm:"type:uuid;not null;index" json:"video_id"`
	Content   string    `gorm:"type:text" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type VideoSegment struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	VideoID        uuid.UUID `gorm:"type:uuid;not null;index" json:"video_id"`
	StartTime      float64   `json:"start_time"`
	EndTime        float64   `json:"end_time"`
	Duration       float64   `json:"duration"`
	Reason         string    `json:"reason"`
	RelevanceScore float64   `json:"relevance_score"`
	ClipURL        string    `json:"clip_url"`
	Status         string    `gorm:"default:'pending'" json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type ProcessingJob struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	VideoID  uuid.UUID `gorm:"type:uuid;not null;index" json:"video_id"`
	Status   string    `gorm:"default:'pending'" json:"status"`
	Progress int       `gorm:"default:0" json:"progress"`
	Error    string    `json:"error"`
	CreatedAt time.Time `json:"created_at"`
}

func (YouTubeVideo) TableName() string {
	return "youtube_videos"
}

func (VideoTranscript) TableName() string {
	return "video_transcripts"
}

func (VideoSegment) TableName() string {
	return "video_segments"
}

func (ProcessingJob) TableName() string {
	return "processing_jobs"
}
