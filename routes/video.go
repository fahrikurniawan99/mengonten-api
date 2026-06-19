package routes

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mengonten-api/models"
	"mengonten-api/utils"
	"mengonten-api/worker"
)

type VideoUploadRequest struct {
	Title string `json:"title" binding:"required"`
}

type ClipJobRequest struct {
	SceneMinGap   float64 `json:"scene_min_gap" binding:"omitempty,min=0.5"`
	MinClipLength float64 `json:"min_clip_length" binding:"omitempty,min=1"`
}

type VideoResponse struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	Duration float64   `json:"duration"`
	Status   string    `json:"status"`
	FileSize int64     `json:"file_size"`
}

type VideoClipResponse struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	StartTime  float64   `json:"start_time"`
	EndTime    float64   `json:"end_time"`
	Duration   float64   `json:"duration"`
	Confidence float64   `json:"confidence"`
	Reason     string    `json:"reason"`
}

// @Summary Upload video
// @Description Upload video file untuk di-analisis dan di-clip
// @Tags Video
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param title formData string true "Video title"
// @Param file formData file true "Video file"
// @Success 201 {object} utils.Response{data=VideoResponse} "Video uploaded successfully"
// @Failure 400 {object} utils.Response "Invalid request"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Router /api/videos/upload [post]
func UploadVideo(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		title := c.PostForm("title")
		if title == "" {
			utils.ErrorResponse(c, http.StatusBadRequest, "Title is required")
			return
		}

		file, err := c.FormFile("file")
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "File is required")
			return
		}

		uploadDir := "uploads/videos"
		filename := fmt.Sprintf("%s_%s", uuid.New().String(), filepath.Base(file.Filename))
		filepath := filepath.Join(uploadDir, filename)

		if err := c.SaveUploadedFile(file, filepath); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save file")
			return
		}

		video := models.Video{
			UserID:   userID.(uuid.UUID),
			Title:    title,
			FilePath: filepath,
			FileSize: file.Size,
			Status:   "uploaded",
		}

		if err := db.Create(&video).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save video record")
			return
		}

		utils.SuccessResponse(c, http.StatusCreated, "Video uploaded successfully", VideoResponse{
			ID:       video.ID,
			Title:    video.Title,
			Status:   video.Status,
			FileSize: video.FileSize,
		})
	}
}

// @Summary Create clip job
// @Description Start auto-clip job untuk video
// @Tags Video
// @Accept json
// @Produce json
// @Security Bearer
// @Param video_id path string true "Video ID"
// @Param request body ClipJobRequest true "Clip job settings"
// @Success 201 {object} utils.Response{data=map[string]string} "Job created"
// @Failure 400 {object} utils.Response "Invalid request"
// @Router /api/videos/{video_id}/clip [post]
func CreateClipJob(db *gorm.DB, clipWorker *worker.ClipWorker) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		videoID := c.Param("video_id")
		parsedVideoID, err := uuid.Parse(videoID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid video ID")
			return
		}

		var video models.Video
		if err := db.Where("id = ? AND user_id = ?", parsedVideoID, userID).First(&video).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Video not found")
			return
		}

		var req ClipJobRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request")
			return
		}

		if req.SceneMinGap == 0 {
			req.SceneMinGap = 2.0
		}
		if req.MinClipLength == 0 {
			req.MinClipLength = 5.0
		}

		clipJob := models.ClipJob{
			VideoID:       parsedVideoID,
			Status:        "pending",
			SceneMinGap:   req.SceneMinGap,
			MinClipLength: req.MinClipLength,
		}

		if err := db.Create(&clipJob).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create job")
			return
		}

		go clipWorker.ProcessClipJob(db, clipJob.ID)

		utils.SuccessResponse(c, http.StatusCreated, "Clip job created", map[string]string{
			"job_id": clipJob.ID.String(),
		})
	}
}

// @Summary Get video clips
// @Description List semua clips dari video
// @Tags Video
// @Produce json
// @Security Bearer
// @Param video_id path string true "Video ID"
// @Success 200 {object} utils.Response{data=[]VideoClipResponse} "Clips retrieved"
// @Failure 404 {object} utils.Response "Video not found"
// @Router /api/videos/{video_id}/clips [get]
func GetVideoClips(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID := c.Param("video_id")
		parsedVideoID, err := uuid.Parse(videoID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid video ID")
			return
		}

		var clips []models.VideoClip
		if err := db.Where("video_id = ?", parsedVideoID).Order("start_time ASC").Find(&clips).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch clips")
			return
		}

		response := make([]VideoClipResponse, len(clips))
		for i, clip := range clips {
			response[i] = VideoClipResponse{
				ID:         clip.ID,
				Title:      clip.Title,
				StartTime:  clip.StartTime,
				EndTime:    clip.EndTime,
				Duration:   clip.Duration,
				Confidence: clip.Confidence,
				Reason:     clip.Reason,
			}
		}

		utils.SuccessResponse(c, http.StatusOK, "Clips retrieved", response)
	}
}

// @Summary Get clip job status
// @Description Check status clip processing job
// @Tags Video
// @Produce json
// @Security Bearer
// @Param job_id path string true "Job ID"
// @Success 200 {object} utils.Response{data=map[string]interface{}} "Job status"
// @Failure 404 {object} utils.Response "Job not found"
// @Router /api/videos/jobs/{job_id} [get]
func GetClipJobStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("job_id")
		parsedJobID, err := uuid.Parse(jobID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid job ID")
			return
		}

		var job models.ClipJob
		if err := db.First(&job, parsedJobID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Job not found")
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "Job status retrieved", map[string]interface{}{
			"job_id":        job.ID,
			"status":        job.Status,
			"progress":      job.Progress,
			"clips_created": job.ClipsCreated,
			"error":         job.ErrorMessage,
		})
	}
}
