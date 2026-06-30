package worker

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type YouTubeMetadata struct {
	Title      string   `json:"title"`
	Duration   float64  `json:"duration"`
	Thumbnail  string   `json:"thumbnail"`
	Tags       []string `json:"tags"`
	Categories []string `json:"categories"`
	IsLive     bool     `json:"is_live"`
}

func ExtractYouTubeMetadata(youtubeURL string) (*YouTubeMetadata, error) {
	cmd := exec.Command("yt-dlp", "--dump-json", "--no-download", youtubeURL)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	var raw struct {
		Title      string   `json:"title"`
		Duration   float64  `json:"duration"`
		Thumbnail  string   `json:"thumbnail"`
		Tags       []string `json:"tags"`
		Categories []string `json:"categories"`
		IsLive     bool     `json:"is_live"`
	}
	if err := json.Unmarshal(output, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &YouTubeMetadata{
		Title:      raw.Title,
		Duration:   raw.Duration,
		Thumbnail:  raw.Thumbnail,
		Tags:       raw.Tags,
		Categories: raw.Categories,
		IsLive:     raw.IsLive,
	}, nil
}
