package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mengonten-api/models"
	"mengonten-api/utils"
	"mengonten-api/worker"
)

type SubmitYouTubeRequest struct {
	YouTubeURL string `json:"youtube_url" binding:"required,url"`
}

type YouTubeVideoResponse struct {
	ID         uuid.UUID `json:"id"`
	YouTubeURL string    `json:"youtube_url"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	Genre      string    `json:"genre"`
}

type VideoSegmentResponse struct {
	ID             uuid.UUID `json:"id"`
	StartTime      float64   `json:"start_time"`
	EndTime        float64   `json:"end_time"`
	Duration       float64   `json:"duration"`
	Reason         string    `json:"reason"`
	RelevanceScore float64   `json:"relevance_score"`
	ClipURL        string    `json:"clip_url"`
	Status         string    `json:"status"`
}

// @Summary Submit YouTube video untuk auto-clipping
// @Description Submit YouTube URL untuk di-download, analyze, dan auto-clip
// @Tags YouTube
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body SubmitYouTubeRequest true "YouTube URL"
// @Success 201 {object} utils.Response{data=YouTubeVideoResponse} "Video submitted"
// @Failure 400 {object} utils.Response "Invalid request"
// @Router /api/youtube/submit [post]
func SubmitYouTubeVideo(db *gorm.DB, processor *worker.YouTubeProcessor) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		var req SubmitYouTubeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
			return
		}

		video := models.YouTubeVideo{
			UserID:     userID.(uuid.UUID),
			YouTubeURL: req.YouTubeURL,
			Status:     "pending",
		}

		if err := db.Create(&video).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to submit video")
			return
		}

		go processor.ProcessYouTubeVideo(db, video.ID)

		utils.SuccessResponse(c, http.StatusCreated, "Video submitted for processing", YouTubeVideoResponse{
			ID:         video.ID,
			YouTubeURL: video.YouTubeURL,
			Status:     video.Status,
		})
	}
}

// @Summary Get video details dan segments
// @Description Get YouTube video info dan list generated clips
// @Tags YouTube
// @Produce json
// @Security Bearer
// @Param video_id path string true "Video ID"
// @Success 200 {object} utils.Response "Video details"
// @Router /api/youtube/{video_id} [get]
func GetYouTubeVideo(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID := c.Param("video_id")
		parsedVideoID, err := uuid.Parse(videoID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid video ID")
			return
		}

		var video models.YouTubeVideo
		if err := db.First(&video, parsedVideoID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Video not found")
			return
		}

		var segments []models.VideoSegment
		db.Where("video_id = ?", parsedVideoID).Find(&segments)

		response := map[string]interface{}{
			"video": YouTubeVideoResponse{
				ID:         video.ID,
				YouTubeURL: video.YouTubeURL,
				Title:      video.Title,
				Status:     video.Status,
				Genre:      video.Genre,
			},
			"segments": segments,
		}

		utils.SuccessResponse(c, http.StatusOK, "Video details retrieved", response)
	}
}

// @Summary Get processing job status
// @Description Check status processing YouTube video
// @Tags YouTube
// @Produce json
// @Security Bearer
// @Param job_id path string true "Job ID"
// @Success 200 {object} utils.Response "Job status"
// @Router /api/youtube/jobs/{job_id} [get]
func GetProcessingJobStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("job_id")
		parsedJobID, err := uuid.Parse(jobID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid job ID")
			return
		}

		var job models.ProcessingJob
		if err := db.First(&job, parsedJobID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Job not found")
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "Job status retrieved", map[string]interface{}{
			"job_id":   job.ID,
			"status":   job.Status,
			"progress": job.Progress,
			"error":    job.Error,
		})
	}
}
