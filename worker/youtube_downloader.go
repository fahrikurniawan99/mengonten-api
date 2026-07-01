package worker

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

type YouTubeDownloader struct {
	OutputDir string
}

func NewYouTubeDownloader(outputDir string) *YouTubeDownloader {
	os.MkdirAll(outputDir, 0755)
	return &YouTubeDownloader{
		OutputDir: outputDir,
	}
}

func (yd *YouTubeDownloader) Download(youtubeURL, outputPath string) error {
	log.Printf("Downloading video from: %s", youtubeURL)

	cmd := exec.Command("yt-dlp",
		"--extractor-args", "youtube:player_client=android,web",
		"-o", outputPath,
		youtubeURL)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		log.Printf("yt-dlp error: %s", errMsg)
		return fmt.Errorf("failed to download video: %s", errMsg)
	}

	if _, err := os.Stat(outputPath); err != nil {
		return fmt.Errorf("downloaded file not found: %v", err)
	}

	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return err
	}

	log.Printf("Video downloaded successfully: %s (size: %d bytes)", outputPath, fileInfo.Size())
	return nil
}