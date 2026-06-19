package worker

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type FrameAnalyzer struct {
	openaiClient *openai.Client
}

type FrameAnalysis struct {
	Timestamp float64
	Interest  float64
	Reason    string
	Emotions  []string
	Activity  string
}

func NewFrameAnalyzer(openaiClient *openai.Client) *FrameAnalyzer {
	return &FrameAnalyzer{
		openaiClient: openaiClient,
	}
}

func (fa *FrameAnalyzer) ExtractFrames(videoPath string, interval float64) ([]string, []float64, error) {
	framePaths := []string{}
	timestamps := []float64{}

	tempDir := fmt.Sprintf("temp_frames_%s", randomString(8))
	os.MkdirAll(tempDir, 0755)

	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vf", fmt.Sprintf("fps=1/%v", interval),
		"-f", "image2",
		fmt.Sprintf("%s/frame_%%04d.png", tempDir))

	if err := cmd.Run(); err != nil {
		return nil, nil, err
	}

	files, err := os.ReadDir(tempDir)
	if err != nil {
		return nil, nil, err
	}

	for i, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".png") {
			framePath := fmt.Sprintf("%s/%s", tempDir, file.Name())
			framePaths = append(framePaths, framePath)
			timestamps = append(timestamps, float64(i)*interval)
		}
	}

	return framePaths, timestamps, nil
}

func (fa *FrameAnalyzer) AnalyzeFrame(ctx context.Context, imagePath string, timestamp float64) (*FrameAnalysis, error) {
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, err
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)

	resp, err := fa.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4VisionPreview,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				Content: `Analyze this video frame and provide:
1. Interest Level (0-100): How interesting/engaging is this frame for content creators?
2. Key Elements: What's happening in the frame?
3. Emotions/Mood detected
4. Activity Level: Idle, Moderate, High

Format your response as JSON:
{
  "interest_level": <number>,
  "activity": "<description>",
  "emotions": ["<emotion1>", "<emotion2>"],
  "reason": "<why this frame is/isn't interesting>"
}

Image data: data:image/png;base64,` + base64Image,
			},
		},
		MaxTokens: 300,
	})

	if err != nil {
		log.Printf("OpenAI API error: %v", err)
		return nil, err
	}

	content := resp.Choices[0].Message.Content

	analysis := &FrameAnalysis{
		Timestamp: timestamp,
		Interest:  parseInterestLevel(content),
		Reason:    content,
		Activity:  "Moderate",
	}

	return analysis, nil
}

func (fa *FrameAnalyzer) AnalyzeMultipleFrames(ctx context.Context, framePaths []string, timestamps []float64) ([]FrameAnalysis, error) {
	analyses := []FrameAnalysis{}

	for i, framePath := range framePaths {
		if i%5 == 0 {
			log.Printf("Analyzing frame %d/%d", i+1, len(framePaths))
		}

		analysis, err := fa.AnalyzeFrame(ctx, framePath, timestamps[i])
		if err != nil {
			log.Printf("Error analyzing frame %d: %v", i, err)
			continue
		}

		analyses = append(analyses, *analysis)
	}

	return analyses, nil
}

func parseInterestLevel(content string) float64 {
	start := strings.Index(content, "\"interest_level\":")
	if start == -1 {
		return 50.0
	}

	start += len("\"interest_level\":")
	end := strings.Index(content[start:], ",")
	if end == -1 {
		end = strings.Index(content[start:], "}")
	}

	valueStr := strings.TrimSpace(content[start : start+end])
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}

	return 50.0
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}

func CleanupFrames(tempDir string) error {
	return os.RemoveAll(tempDir)
}
