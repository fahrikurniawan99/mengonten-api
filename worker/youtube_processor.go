package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"mengonten-api/config"
	"mengonten-api/models"
)

type YouTubeProcessor struct {
	ExternalAPIs     *config.ExternalAPIs
	TempDir          string
	OutputDir        string
	CloudinaryUpload bool
}

func NewYouTubeProcessor(externalAPIs *config.ExternalAPIs) *YouTubeProcessor {
	tempDir := filepath.Join(os.TempDir(), "youtube_clips")
	outputDir := "uploads/youtube_clips"

	os.MkdirAll(tempDir, 0755)
	os.MkdirAll(outputDir, 0755)

	return &YouTubeProcessor{
		ExternalAPIs:     externalAPIs,
		TempDir:          tempDir,
		OutputDir:        outputDir,
		CloudinaryUpload: externalAPIs.CloudinaryName != "",
	}
}

func (yp *YouTubeProcessor) ProcessYouTubeVideo(db *gorm.DB, videoID uuid.UUID) {
	var video models.YouTubeVideo
	if err := db.First(&video, videoID).Error; err != nil {
		log.Printf("Video not found: %v", err)
		return
	}

	job := models.ProcessingJob{
		VideoID: videoID,
		Status:  "downloading",
	}
	db.Create(&job)

	video.Status = "downloading"
	db.Model(&video).Update("status", "downloading")

	if err := yp.downloadVideo(db, &video, &job); err != nil {
		yp.updateJobError(db, &job, video, "download", err.Error())
		return
	}

	video.Status = "transcribing"
	db.Model(&video).Update("status", "transcribing")
	job.Status = "transcribing"
	job.Progress = 20
	db.Model(&job).Updates(job)

	transcript, err := yp.generateTranscript(db, &video, &job)
	if err != nil {
		yp.updateJobError(db, &job, video, "transcribe", err.Error())
		return
	}

	video.Status = "analyzing"
	db.Model(&video).Update("status", "analyzing")
	job.Status = "analyzing"
	job.Progress = 40
	db.Model(&job).Updates(job)

	genre, err := yp.analyzeGenre(db, &video, transcript)
	if err != nil {
		yp.updateJobError(db, &job, video, "genre_analysis", err.Error())
		return
	}

	video.Genre = genre
	db.Model(&video).Update("genre", genre)

	video.Status = "segmenting"
	db.Model(&video).Update("status", "segmenting")
	job.Status = "segmenting"
	job.Progress = 60
	db.Model(&job).Updates(job)

	segments, err := yp.findInterestingSegments(db, &video, transcript, genre)
	if err != nil {
		yp.updateJobError(db, &job, video, "segmenting", err.Error())
		return
	}

	video.Status = "clipping"
	db.Model(&video).Update("status", "clipping")
	job.Status = "clipping"
	job.Progress = 75
	db.Model(&job).Updates(job)

	if err := yp.createAndUploadClips(db, &video, &job, segments); err != nil {
		yp.updateJobError(db, &job, video, "clipping", err.Error())
		return
	}

	video.Status = "completed"
	video.ProcessedAt = timePtr(time.Now())
	db.Model(&video).Updates(video)

	job.Status = "completed"
	job.Progress = 100
	db.Model(&job).Updates(job)

	log.Printf("YouTube video %s processing completed successfully", video.ID)
}

func (yp *YouTubeProcessor) downloadVideo(db *gorm.DB, video *models.YouTubeVideo, job *models.ProcessingJob) error {
	videoPath := filepath.Join(yp.TempDir, fmt.Sprintf("%s.mp4", video.ID.String()))

	downloader := NewYouTubeDownloader(yp.TempDir)
	if err := downloader.Download(video.YouTubeURL, videoPath); err != nil {
		return err
	}

	video.LocalFilePath = videoPath
	db.Model(video).Update("local_file_path", videoPath)

	job.Progress = 15
	db.Model(job).Update("progress", 15)

	return nil
}

func (yp *YouTubeProcessor) generateTranscript(db *gorm.DB, video *models.YouTubeVideo, job *models.ProcessingJob) (string, error) {
	generator := NewTranscriptGenerator(yp.ExternalAPIs.WhisperAPIKey, yp.TempDir)

	transcript, err := generator.GenerateTranscript(context.Background(), video.LocalFilePath)
	if err != nil {
		return "", err
	}

	dbTranscript := models.VideoTranscript{
		VideoID: video.ID,
		Content: transcript,
	}
	db.Create(&dbTranscript)

	job.Progress = 35
	db.Model(job).Update("progress", 35)

	return transcript, nil
}

func (yp *YouTubeProcessor) analyzeGenre(db *gorm.DB, video *models.YouTubeVideo, transcript string) (string, error) {
	analyzer := NewGenreAnalyzer(yp.ExternalAPIs.GPTAPIKey)

	genre, err := analyzer.AnalyzeGenre(context.Background(), video.YouTubeURL, transcript)
	if err != nil {
		return "", err
	}

	return genre, nil
}

func (yp *YouTubeProcessor) findInterestingSegments(db *gorm.DB, video *models.YouTubeVideo, transcript, genre string) ([]models.VideoSegment, error) {
	analyzer := NewGenreAnalyzer(yp.ExternalAPIs.GPTAPIKey)

	segments, err := analyzer.FindInterestingSegments(context.Background(), transcript, genre)
	if err != nil {
		return nil, err
	}

	return segments, nil
}

func (yp *YouTubeProcessor) createAndUploadClips(db *gorm.DB, video *models.YouTubeVideo, job *models.ProcessingJob, segments []models.VideoSegment) error {
	for i, segment := range segments {
		clipPath := filepath.Join(yp.OutputDir, fmt.Sprintf("%s_clip_%d.mp4", video.ID.String(), i+1))

		clipper := NewVideoClipper()
		if err := clipper.CutVideo(video.LocalFilePath, clipPath, segment.StartTime, segment.EndTime); err != nil {
			log.Printf("Failed to cut clip %d: %v", i+1, err)
			continue
		}

		segment.Status = "created"

		if yp.CloudinaryUpload {
			uploader := NewCloudinaryUploader(yp.ExternalAPIs)
			clipURL, err := uploader.Upload(clipPath, fmt.Sprintf("youtube_clips/%s", video.ID.String()))
			if err != nil {
				log.Printf("Failed to upload clip %d to Cloudinary: %v", i+1, err)
				segment.Status = "created_local"
			} else {
				segment.ClipURL = clipURL
				segment.Status = "uploaded"
			}
		}

		segment.VideoID = video.ID
		db.Create(&segment)

		job.Progress = 75 + (25 * (i + 1) / len(segments))
		db.Model(job).Update("progress", job.Progress)
	}

	return nil
}

func (yp *YouTubeProcessor) updateJobError(db *gorm.DB, job *models.ProcessingJob, video models.YouTubeVideo, stage string, errMsg string) {
	fullError := fmt.Sprintf("Error in %s stage: %s", stage, errMsg)

	job.Status = "failed"
	job.Error = fullError
	db.Model(job).Updates(job)

	video.Status = "failed"
	db.Model(&video).Update("status", "failed")

	log.Printf("YouTube processing job %s failed: %s", job.ID, fullError)
}

func timePtr(t time.Time) *time.Time {
	return &t
}
