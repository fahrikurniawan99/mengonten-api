package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type TranscriptGenerator struct {
	APIKey  string
	TempDir string
	Client  *openai.Client
}

func NewTranscriptGenerator(apiKey string, tempDir string) *TranscriptGenerator {
	return &TranscriptGenerator{
		APIKey:  apiKey,
		TempDir: tempDir,
		Client:  openai.NewClient(apiKey),
	}
}

func (tg *TranscriptGenerator) GenerateTranscript(ctx context.Context, videoPath string) (string, error) {
	log.Printf("Processing file for transcription: %s", videoPath)

	audioPath := videoPath
	ext := strings.ToLower(filepath.Ext(videoPath))

	// Check if file is already audio format
	audioFormats := map[string]bool{
		".mp3": true, ".wav": true, ".m4a": true, ".aac": true,
		".flac": true, ".ogg": true, ".wma": true,
	}

	if !audioFormats[ext] {
		// Need to extract audio from video
		log.Printf("Extracting audio from video: %s", videoPath)
		audioPath = filepath.Join(tg.TempDir, "audio.mp3")
		if err := tg.extractAudio(videoPath, audioPath); err != nil {
			return "", fmt.Errorf("failed to extract audio: %v", err)
		}
		defer os.Remove(audioPath)
	} else {
		log.Printf("File is already audio format, skipping extraction: %s", videoPath)
	}

	log.Printf("Generating transcript using Whisper API")

	audioFile, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %v", err)
	}
	defer audioFile.Close()

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: audioPath,
	}

	resp, err := tg.Client.CreateTranscription(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("whisper API error: %v", err)
	}

	log.Printf("Transcript generated successfully. Length: %d characters", len(resp.Text))

	return resp.Text, nil
}

func (tg *TranscriptGenerator) extractAudio(videoPath, audioPath string) error {
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-q:a", "9",
		"-n",
		audioPath)

	if err := cmd.Run(); err != nil {
		return err
	}

	if _, err := os.Stat(audioPath); err != nil {
		return fmt.Errorf("audio file not created: %v", err)
	}

	return nil
}
