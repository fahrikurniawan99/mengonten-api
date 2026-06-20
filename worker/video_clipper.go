package worker

import (
	"fmt"
	"log"
	"os/exec"
)

type VideoClipper struct{}

func NewVideoClipper() *VideoClipper {
	return &VideoClipper{}
}

func (vc *VideoClipper) CutVideo(inputPath, outputPath string, startTime, endTime float64) error {
	log.Printf("Cutting video from %.2f to %.2f", startTime, endTime)

	duration := endTime - startTime

	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-ss", fmt.Sprintf("%.2f", startTime),
		"-t", fmt.Sprintf("%.2f", duration),
		"-c:v", "libx264",
		"-c:a", "aac",
		"-y",
		outputPath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg cut error: %v", err)
	}

	log.Printf("Video clip created successfully: %s", outputPath)
	return nil
}
