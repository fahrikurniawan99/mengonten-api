package worker

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
		"-f", "best",
		"-o", outputPath,
		youtubeURL)

	if err := cmd.Run(); err != nil {
		log.Printf("yt-dlp error: %v", err)
		return fmt.Errorf("failed to download video: %v", err)
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
