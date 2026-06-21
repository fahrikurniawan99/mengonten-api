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

func (vc *VideoClipper) CutVideo(inputPath, outputPath string, startTime, endTime float64, quality string) error {
	log.Printf("Cutting video from %.2f to %.2f (quality: %s)", startTime, endTime, quality)

	duration := endTime - startTime

	args := []string{
		"-i", inputPath,
		"-ss", fmt.Sprintf("%.2f", startTime),
		"-t", fmt.Sprintf("%.2f", duration),
	}

	if quality != "" && quality != "original" {
		resolution := getResolution(quality)
		if resolution != "" {
			args = append(args, "-vf", resolution)
		}
	}

	args = append(args,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-y",
		outputPath)

	cmd := exec.Command("ffmpeg", args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg cut error: %v", err)
	}

	log.Printf("Video clip created successfully: %s", outputPath)
	return nil
}

func getResolution(quality string) string {
	switch quality {
	case "720p":
		return "scale=-1:720"
	case "1080p":
		return "scale=-1:1080"
	case "480p":
		return "scale=-1:480"
	default:
		return ""
	}
}
