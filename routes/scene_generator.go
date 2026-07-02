package routes

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"mengonten-api/models"
	"mengonten-api/utils"
	"mengonten-api/worker"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateSceneRequest struct {
	YouTubeURL string `json:"youtube_url" binding:"required,url"`
}

// @Summary Create chapter segmentation job
// @Description Submit YouTube URL untuk di-analyze jadi chapters
// @Tags Scene
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateSceneRequest true "YouTube URL"
// @Success 202 {object} utils.Response "Job created"
// @Router /api/youtube/scenes [post]
func CreateSceneJob(db *gorm.DB, segmenter *worker.ChapterSegmenter, transcriptGen *worker.TranscriptGenerator) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		var req CreateSceneRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request")
			return
		}

		job := models.SceneJob{
			UserID:     userID.(uuid.UUID),
			YouTubeURL: req.YouTubeURL,
			Status:     "pending",
			Progress:   0,
		}

		if err := db.Create(&job).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create job")
			return
		}

		go processSceneJob(db, job.ID, segmenter, transcriptGen)

		utils.SuccessResponse(c, http.StatusAccepted, "Job created", gin.H{
			"job_id": job.ID,
			"status": "pending",
		})
	}
}

// @Summary Get scene job status
// @Description Get job status dan hasil chapters
// @Tags Scene
// @Produce json
// @Security Bearer
// @Param job_id path string true "Job ID"
// @Success 200 {object} utils.Response "Job status"
// @Router /api/youtube/scenes/{job_id} [get]
func GetSceneJob(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		jobID := c.Param("job_id")
		parsedID, err := uuid.Parse(jobID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid job ID")
			return
		}

		var job models.SceneJob
		if err := db.Where("id = ? AND user_id = ?", parsedID, userID).First(&job).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Job not found")
			return
		}

		response := gin.H{
			"job_id":     job.ID,
			"status":     job.Status,
			"progress":   job.Progress,
			"created_at": job.CreatedAt,
			"updated_at": job.UpdatedAt,
		}

		if job.Status == "completed" {
			var chapters []worker.ChapterData
			if err := json.Unmarshal(job.Chapters, &chapters); err == nil {
				response["chapters"] = chapters
			}
			response["transcript"] = job.Transcript
		}

		if job.Error != "" {
			response["error"] = job.Error
		}

		utils.SuccessResponse(c, http.StatusOK, "Job retrieved", response)
	}
}

// @Summary List scene jobs (user)
// @Description List semua scene jobs milik user dengan pagination
// @Tags Scene
// @Produce json
// @Security Bearer
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 20)"
// @Param status query string false "Filter by status (pending/processing/completed/failed)"
// @Success 200 {object} utils.Response "Jobs list"
// @Router /api/youtube/scenes [get]
func ListSceneJobs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		page := 1
		if p := c.Query("page"); p != "" {
			if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
				page = parsed
			}
		}

		limit := 20
		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
				limit = parsed
			}
		}

		offset := (page - 1) * limit
		status := c.Query("status")

		var jobs []models.SceneJob
		query := db.Where("user_id = ?", userID)

		if status != "" {
			query = query.Where("status = ?", status)
		}

		var total int64
		query.Model(&models.SceneJob{}).Count(&total)

		if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&jobs).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch jobs")
			return
		}

		data := make([]gin.H, 0)
		for _, job := range jobs {
			chaptersCount := 0
			if len(job.Chapters) > 0 {
				var chapters []worker.ChapterData
				if err := json.Unmarshal(job.Chapters, &chapters); err == nil {
					chaptersCount = len(chapters)
				}
			}

			var tags []string
			var categories []string
			if len(job.Tags) > 0 {
				json.Unmarshal(job.Tags, &tags)
			}
			if len(job.Categories) > 0 {
				json.Unmarshal(job.Categories, &categories)
			}

			data = append(data, gin.H{
				"id":             job.ID,
				"youtube_url":    job.YouTubeURL,
				"title":          job.Title,
				"duration":       job.Duration,
				"thumbnail":      job.Thumbnail,
				"tags":           tags,
				"categories":     categories,
				"is_live":        job.IsLive,
				"status":         job.Status,
				"progress":       job.Progress,
				"chapters_count": chaptersCount,
				"created_at":     job.CreatedAt,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Jobs retrieved",
			"page":    page,
			"limit":   limit,
			"total":   total,
			"data":    data,
		})
	}
}

func processSceneJob(db *gorm.DB, jobID uuid.UUID, segmenter *worker.ChapterSegmenter, transcriptGen *worker.TranscriptGenerator) {
	log.Printf("[SceneJob] Starting processing for job %s", jobID.String())
	tmpDir, err := os.MkdirTemp("", "scene_job_*")
	if err != nil {
		log.Printf("[SceneJob] Failed to create temp dir: %v", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	var job models.SceneJob
	if err := db.First(&job, jobID).Error; err != nil {
		log.Printf("[SceneJob] Job not found: %v", err)
		return
	}

	if err := db.Model(&job).Updates(map[string]interface{}{"status": "processing", "progress": 10}).Error; err != nil {
		log.Printf("[SceneJob] Failed to update job status: %v", err)
		return
	}

	metadata, err := worker.ExtractYouTubeMetadata(job.YouTubeURL)
	if err == nil {
		tagsJSON, _ := json.Marshal(metadata.Tags)
		categoriesJSON, _ := json.Marshal(metadata.Categories)
		db.Model(&job).Updates(map[string]interface{}{
			"title":      metadata.Title,
			"duration":   metadata.Duration,
			"thumbnail":  metadata.Thumbnail,
			"tags":       tagsJSON,
			"categories": categoriesJSON,
			"is_live":    metadata.IsLive,
		})
	}

	db.Model(&job).Update("progress", 20)

	downloader := worker.NewYouTubeDownloader(tmpDir)
	audioPath := tmpDir + "/audio.mp3"

	if err := downloader.Download(job.YouTubeURL, audioPath); err != nil {
		log.Printf("[SceneJob] Download failed: %v", err)
		db.Model(&job).Updates(map[string]interface{}{
			"status": "failed",
			"error":  "Download failed: " + err.Error(),
		})
		return
	}

	db.Model(&job).Update("progress", 40)

	log.Printf("[SceneJob] Downloaded audio to: %s", audioPath)

	txnCtx, txnCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	transcript, err := transcriptGen.GenerateTranscriptChunked(txnCtx, audioPath)
	txnCancel()
	if err != nil {
		log.Printf("[SceneJob] Transcription failed: %v", err)
		db.Model(&job).Updates(map[string]interface{}{
			"status": "failed",
			"error":  "Transcription failed: " + err.Error(),
		})
		return
	}

	log.Printf("[SceneJob] Transcription completed, length: %d characters", len(transcript))

	// db.Model(&job).Updates(map[string]interface{}{"transcript": transcript, "progress": 60})

	// groqCtx, groqCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	// chapters, err := segmenter.SegmentChapters(groqCtx, transcript)
	// groqCancel()
	// if err != nil {
	// 	log.Printf("[SceneJob] Segmentation failed: %v", err)
	// 	db.Model(&job).Updates(map[string]interface{}{
	// 		"status": "failed",
	// 		"error":  "Segmentation failed: " + err.Error(),
	// 	})
	// 	return
	// }

	// chaptersJSON, err := json.Marshal(chapters)
	// if err != nil {
	// 	log.Printf("[SceneJob] Failed to marshal chapters: %v", err)
	// 	db.Model(&job).Updates(map[string]interface{}{
	// 		"status": "failed",
	// 		"error":  "Failed to marshal chapters: " + err.Error(),
	// 	})
	// 	return
	// }
	// db.Model(&job).Updates(map[string]interface{}{
	// 	"chapters": driver.Value(chaptersJSON),
	// 	"status":   "completed",
	// 	"progress": 100,
	// })

	// log.Printf("[SceneJob] Completed: %d chapters generated", len(chapters))
}
