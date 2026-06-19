package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
	"mengonten-api/models"
)

type ClipWorker struct {
	FFmpegPath   string
	OutputDir    string
	OpenAIClient *openai.Client
}

type SceneFrame struct {
	Timestamp float64
	Score     float64
}

func NewClipWorker(ffmpegPath, outputDir string) *ClipWorker {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	if outputDir == "" {
		outputDir = "uploads/clips"
	}

	os.MkdirAll(outputDir, 0755)

	return &ClipWorker{
		FFmpegPath: ffmpegPath,
		OutputDir:  outputDir,
	}
}

func (cw *ClipWorker) SetOpenAIClient(client *openai.Client) {
	cw.OpenAIClient = client
}

func (cw *ClipWorker) ProcessClipJob(db *gorm.DB, jobID uuid.UUID) {
	var job models.ClipJob
	if err := db.First(&job, jobID).Error; err != nil {
		log.Printf("Job not found: %v", err)
		return
	}

	db.Model(&job).Update("status", "processing")

	var video models.Video
	if err := db.First(&video, job.VideoID).Error; err != nil {
		cw.updateJobError(db, &job, "Video not found")
		return
	}

	if cw.OpenAIClient != nil {
		cw.processClipJobWithOpenAI(db, &job, &video)
	} else {
		cw.processClipJobWithSceneDetection(db, &job, &video)
	}
}

func (cw *ClipWorker) processClipJobWithOpenAI(db *gorm.DB, job *models.ClipJob, video *models.Video) {
	duration, err := cw.getVideoDuration(video.FilePath)
	if err != nil {
		cw.updateJobError(db, job, fmt.Sprintf("Failed to get duration: %v", err))
		return
	}
	db.Model(video).Update("duration", duration)

	analyzer := NewFrameAnalyzer(cw.OpenAIClient)
	framePaths, timestamps, err := analyzer.ExtractFrames(video.FilePath, 5.0)
	if err != nil {
		cw.updateJobError(db, job, fmt.Sprintf("Failed to extract frames: %v", err))
		return
	}
	defer CleanupFrames("temp_frames_" + framePaths[0][:8])

	ctx := context.Background()
	analyses, err := analyzer.AnalyzeMultipleFrames(ctx, framePaths, timestamps)
	if err != nil {
		cw.updateJobError(db, job, fmt.Sprintf("OpenAI analysis failed: %v", err))
		return
	}

	clips := cw.generateClipsFromAnalysis(analyses, duration, job.MinClipLength)
	if err := cw.createClipFiles(db, job, video, clips); err != nil {
		cw.updateJobError(db, job, fmt.Sprintf("Failed to create clips: %v", err))
		return
	}

	db.Model(job).Updates(map[string]interface{}{
		"status":        "completed",
		"clips_created": len(clips),
		"progress":      100,
	})
}

func (cw *ClipWorker) processClipJobWithSceneDetection(db *gorm.DB, job *models.ClipJob, video *models.Video) {
	duration, err := cw.getVideoDuration(video.FilePath)
	if err != nil {
		cw.updateJobError(db, job, fmt.Sprintf("Failed to get video duration: %v", err))
		return
	}
	db.Model(video).Update("duration", duration)

	sceneFrames, err := cw.detectScenes(video.FilePath, job.SceneMinGap)
	if err != nil {
		cw.updateJobError(db, job, fmt.Sprintf("Scene detection failed: %v", err))
		return
	}

	clips := cw.generateClipsFromScenes(sceneFrames, duration, job.MinClipLength)
	if err := cw.createClipFiles(db, job, video, clips); err != nil {
		cw.updateJobError(db, job, fmt.Sprintf("Failed to create clips: %v", err))
		return
	}

	db.Model(job).Updates(map[string]interface{}{
		"status":        "completed",
		"clips_created": len(clips),
		"progress":      100,
	})
}

func (cw *ClipWorker) getVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command(cw.FFmpegPath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1:noprint_wrappers=1",
		videoPath)

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, err
	}

	return duration, nil
}

func (cw *ClipWorker) detectScenes(videoPath string, minGap float64) ([]SceneFrame, error) {
	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("scenedetect_%s.txt", uuid.New().String()))
	defer os.Remove(outputPath)

	cmd := exec.Command(cw.FFmpegPath,
		"-i", videoPath,
		"-vf", "select='gt(scene\\,.4)',metadata=print:file="+outputPath,
		"-vsync", "vfr",
		"-f", "null",
		"-")

	if err := cmd.Run(); err != nil {
		log.Printf("FFmpeg scene detection warning: %v", err)
	}

	sceneFrames := []SceneFrame{}
	content, err := os.ReadFile(outputPath)
	if err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.Contains(line, "pts_time") {
				parts := strings.Split(line, "=")
				if len(parts) >= 2 {
					timeStr := strings.TrimSpace(parts[len(parts)-1])
					if timestamp, err := strconv.ParseFloat(timeStr, 64); err == nil {
						if len(sceneFrames) == 0 || timestamp-sceneFrames[len(sceneFrames)-1].Timestamp >= minGap {
							sceneFrames = append(sceneFrames, SceneFrame{
								Timestamp: timestamp,
								Score:     0.8,
							})
						}
					}
				}
			}
		}
	}

	if len(sceneFrames) == 0 {
		sceneFrames = append(sceneFrames, SceneFrame{Timestamp: 0, Score: 1.0})
	}

	return sceneFrames, nil
}

func (cw *ClipWorker) generateClipsFromAnalysis(analyses []FrameAnalysis, totalDuration float64, minLength float64) []models.VideoClip {
	clips := []models.VideoClip{}
	var currentClip *models.VideoClip

	for _, analysis := range analyses {
		if analysis.Interest >= 70.0 {
			if currentClip == nil {
				currentClip = &models.VideoClip{
					StartTime:  analysis.Timestamp,
					Confidence: analysis.Interest / 100.0,
					Reason:     analysis.Reason,
				}
			}
			currentClip.EndTime = analysis.Timestamp
		} else if currentClip != nil {
			duration := currentClip.EndTime - currentClip.StartTime
			if duration >= minLength {
				currentClip.Duration = duration
				clips = append(clips, *currentClip)
			}
			currentClip = nil
		}
	}

	if currentClip != nil {
		duration := currentClip.EndTime - currentClip.StartTime
		if duration >= minLength {
			currentClip.Duration = duration
			clips = append(clips, *currentClip)
		}
	}

	return clips
}

func (cw *ClipWorker) generateClipsFromScenes(scenes []SceneFrame, totalDuration float64, minLength float64) []models.VideoClip {
	clips := []models.VideoClip{}

	for i := 0; i < len(scenes)-1; i++ {
		startTime := scenes[i].Timestamp
		endTime := scenes[i+1].Timestamp

		duration := endTime - startTime
		if duration >= minLength {
			clip := models.VideoClip{
				StartTime:  startTime,
				EndTime:    endTime,
				Duration:   duration,
				Confidence: (scenes[i].Score + scenes[i+1].Score) / 2,
				Reason:     fmt.Sprintf("Scene detected at %.2fs-%.2fs", startTime, endTime),
			}
			clips = append(clips, clip)
		}
	}

	if len(scenes) > 0 {
		lastScene := scenes[len(scenes)-1]
		if totalDuration-lastScene.Timestamp >= minLength {
			clip := models.VideoClip{
				StartTime:  lastScene.Timestamp,
				EndTime:    totalDuration,
				Duration:   totalDuration - lastScene.Timestamp,
				Confidence: lastScene.Score,
				Reason:     fmt.Sprintf("Final scene at %.2fs", lastScene.Timestamp),
			}
			clips = append(clips, clip)
		}
	}

	return clips
}

func (cw *ClipWorker) createClipFiles(db *gorm.DB, job *models.ClipJob, video *models.Video, clips []models.VideoClip) error {
	for i, clip := range clips {
		clipID := uuid.New()
		outputFilename := fmt.Sprintf("%s_clip_%d.mp4", video.ID.String(), i+1)
		outputPath := filepath.Join(cw.OutputDir, outputFilename)

		cmd := exec.Command(cw.FFmpegPath,
			"-i", video.FilePath,
			"-ss", fmt.Sprintf("%.2f", clip.StartTime),
			"-to", fmt.Sprintf("%.2f", clip.EndTime),
			"-c:v", "libx264",
			"-c:a", "aac",
			"-y",
			outputPath)

		if err := cmd.Run(); err != nil {
			log.Printf("Failed to create clip %d: %v", i+1, err)
			continue
		}

		clip.ID = clipID
		clip.VideoID = video.ID
		clip.FilePath = outputPath
		clip.Title = fmt.Sprintf("%s - Clip %d", video.Title, i+1)

		if err := db.Create(&clip).Error; err != nil {
			log.Printf("Failed to save clip record: %v", err)
			continue
		}

		job.ClipsCreated++
		job.Progress = int((float64(i+1) / float64(len(clips))) * 100)
		db.Model(&job).Updates(map[string]interface{}{
			"clips_created": job.ClipsCreated,
			"progress":      job.Progress,
		})
	}

	return nil
}

func (cw *ClipWorker) updateJobError(db *gorm.DB, job *models.ClipJob, errMsg string) {
	db.Model(job).Updates(map[string]interface{}{
		"status":         "failed",
		"error_message":  errMsg,
	})
	log.Printf("Clip job %s failed: %s", job.ID, errMsg)
}
