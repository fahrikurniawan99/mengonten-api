package routes

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mengonten-api/config"
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

		rules := GetUserSubscriptionRules(db, userID.(uuid.UUID))
		if rules == nil {
			utils.ErrorResponse(c, http.StatusForbidden, "No active subscription. Please subscribe first.")
			return
		}
		if maxStorageStr, ok := rules["max_storage_mb"]; ok {
			var maxStorageMB int
			fmt.Sscanf(maxStorageStr, "%d", &maxStorageMB)

			var subscription models.UserSubscription
			if err := db.Where("user_id = ? AND status = ?", userID, "active").
				Order("end_date DESC").First(&subscription).Error; err == nil {
				usedMB := float64(subscription.StorageUsedBytes) / (1024 * 1024)
				if maxStorageMB > 0 && int(usedMB) >= maxStorageMB {
					utils.ErrorResponse(c, http.StatusForbidden,
						fmt.Sprintf("Storage limit reached (%dMB / %dMB). Please wait or upgrade your plan.", int(usedMB), maxStorageMB))
					return
				}
			}
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

		var video models.YouTubeVideo
		if err := db.First(&video, parsedVideoID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Video not found")
			return
		}

		if video.UserID != userID.(uuid.UUID) {
			utils.ErrorResponse(c, http.StatusForbidden, "Access denied")
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
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

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

		var video models.YouTubeVideo
		if err := db.First(&video, job.VideoID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Video not found")
			return
		}

		if video.UserID != userID.(uuid.UUID) {
			utils.ErrorResponse(c, http.StatusForbidden, "Access denied")
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

// @Summary Delete video segment/clip
// @Description Delete segment dari database dan Cloudinary
// @Tags YouTube
// @Produce json
// @Security Bearer
// @Param segment_id path string true "Segment ID"
// @Success 200 {object} utils.Response "Segment deleted"
// @Failure 404 {object} utils.Response "Segment not found"
// @Router /api/youtube/segments/{segment_id} [delete]
func DeleteVideoSegment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		segmentID := c.Param("segment_id")
		parsedSegmentID, err := uuid.Parse(segmentID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid segment ID")
			return
		}

		var segment models.VideoSegment
		if err := db.First(&segment, parsedSegmentID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Segment not found")
			return
		}

		var video models.YouTubeVideo
		if err := db.First(&video, segment.VideoID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Video not found")
			return
		}

		if video.UserID != userID.(uuid.UUID) {
			utils.ErrorResponse(c, http.StatusForbidden, "Access denied")
			return
		}

		if segment.ClipURL != "" {
			if err := deleteFromCloudinary(segment.ClipURL); err != nil {
				log.Printf("Failed to delete from Cloudinary: %v", err)
			} else {
				log.Printf("Deleted from Cloudinary: %s", segment.ClipURL)
			}
		}

		if err := db.Delete(&segment).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete segment")
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "Segment deleted successfully", nil)
	}
}

func deleteFromCloudinary(url string) error {
	if url == "" {
		return nil
	}

	externalAPIs := config.InitExternalAPIs()
	if externalAPIs.R2AccountID == "" {
		return fmt.Errorf("R2 not configured")
	}

	publicURL := externalAPIs.R2PublicURL
	publicURL = strings.TrimRight(publicURL, "/")

	if !strings.HasPrefix(url, publicURL) {
		return fmt.Errorf("URL does not belong to R2 storage")
	}

	key := strings.TrimPrefix(url, publicURL+"/")
	key = strings.TrimPrefix(key, "/")

	return worker.DeleteFromR2(key)
}

type VideoListResponse struct {
	ID          uuid.UUID `json:"id"`
	YouTubeURL string    `json:"youtube_url"`
	Title       string    `json:"title"`
	Genre       string    `json:"genre"`
	Status      string    `json:"status"`
	Duration    float64   `json:"duration"`
	ClipsCount  int       `json:"clips_count"`
	CreatedAt   string    `json:"created_at"`
}

// @Summary List my videos
// @Description List video milik user sendiri dengan pagination
// @Tags YouTube
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status"
// @Param search query string false "Search by title/genre"
// @Success 200 {object} utils.Response "Video list with pagination"
// @Router /api/youtube [get]
func GetMyVideos(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		status := c.Query("status")
		search := c.Query("search")

		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 10
		}

		query := db.Where("user_id = ?", userID)

		if status != "" {
			query = query.Where("status = ?", status)
		}
		if search != "" {
			query = query.Where("LOWER(title) LIKE ? OR LOWER(genre) LIKE ?",
				"%"+strings.ToLower(search)+"%",
				"%"+strings.ToLower(search)+"%")
		}

		var total int64
		query.Model(&models.YouTubeVideo{}).Count(&total)

		var videos []models.YouTubeVideo
		offset := (page - 1) * limit
		if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&videos).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch videos")
			return
		}

		response := make([]VideoListResponse, len(videos))
		for i, v := range videos {
			var clipsCount int64
			db.Model(&models.VideoSegment{}).Where("video_id = ?", v.ID).Count(&clipsCount)

			response[i] = VideoListResponse{
				ID:          v.ID,
				YouTubeURL:  v.YouTubeURL,
				Title:       v.Title,
				Genre:       v.Genre,
				Status:      v.Status,
				Duration:    v.Duration,
				ClipsCount:  int(clipsCount),
				CreatedAt:   v.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}

		totalPages := int(total) / limit
		if int(total)%limit != 0 {
			totalPages++
		}

		utils.SuccessResponse(c, http.StatusOK, "Videos retrieved", map[string]interface{}{
			"videos": response,
			"pagination": map[string]interface{}{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
			},
		})
	}
}

// @Summary List all videos (admin)
// @Description List semua video dari semua user
// @Tags Admin
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status"
// @Param search query string false "Search by title/genre"
// @Success 200 {object} utils.Response "Video list with pagination"
// @Router /api/admin/videos [get]
func GetAllVideos(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		status := c.Query("status")
		search := c.Query("search")

		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 10
		}

		query := db.Model(&models.YouTubeVideo{})

		if status != "" {
			query = query.Where("status = ?", status)
		}
		if search != "" {
			query = query.Where("LOWER(title) LIKE ? OR LOWER(genre) LIKE ?",
				"%"+strings.ToLower(search)+"%",
				"%"+strings.ToLower(search)+"%")
		}

		var total int64
		query.Count(&total)

		var videos []models.YouTubeVideo
		offset := (page - 1) * limit
		if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&videos).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch videos")
			return
		}

		response := make([]VideoListResponse, len(videos))
		for i, v := range videos {
			var clipsCount int64
			db.Model(&models.VideoSegment{}).Where("video_id = ?", v.ID).Count(&clipsCount)

			response[i] = VideoListResponse{
				ID:          v.ID,
				YouTubeURL:  v.YouTubeURL,
				Title:       v.Title,
				Genre:       v.Genre,
				Status:      v.Status,
				Duration:    v.Duration,
				ClipsCount:  int(clipsCount),
				CreatedAt:   v.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}

		totalPages := int(total) / limit
		if int(total)%limit != 0 {
			totalPages++
		}

		utils.SuccessResponse(c, http.StatusOK, "Videos retrieved", map[string]interface{}{
			"videos": response,
			"pagination": map[string]interface{}{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
			},
		})
	}
}
