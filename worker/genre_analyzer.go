package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/sashabaranov/go-openai"
	"mengonten-api/models"
)

type GenreAnalyzer struct {
	APIKey string
	Client *openai.Client
}

type GenreAnalysisResponse struct {
	Genre       string  `json:"genre"`
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
}

type SegmentFinderResponse struct {
	Segments []SegmentInfo `json:"segments"`
}

type SegmentInfo struct {
	StartTime      float64 `json:"start_time"`
	EndTime        float64 `json:"end_time"`
	Reason         string  `json:"reason"`
	RelevanceScore float64 `json:"relevance_score"`
	Keyword        string  `json:"keyword"`
}

func NewGenreAnalyzer(apiKey string) *GenreAnalyzer {
	return &GenreAnalyzer{
		APIKey: apiKey,
		Client: openai.NewClient(apiKey),
	}
}

func (ga *GenreAnalyzer) AnalyzeGenre(ctx context.Context, videoURL, transcript string) (string, error) {
	log.Printf("Analyzing video genre from transcript")

	prompt := fmt.Sprintf(`Analyze this video transcript and determine the primary genre.
Possible genres: Comedy, Educational, Entertainment, Sports, Music, News, Gaming, Vlog, Tutorial, Documentary, Drama, Animation, Other

Video URL: %s
Transcript excerpt (first 2000 chars): %s

Respond in JSON format:
{
  "genre": "genre_name",
  "confidence": 0.0-1.0,
  "description": "brief description"
}`, videoURL, truncate(transcript, 2000))

	resp, err := ga.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.3,
		MaxTokens:   200,
	})

	if err != nil {
		return "", fmt.Errorf("GPT API error: %v", err)
	}

	content := resp.Choices[0].Message.Content
	var analysisResp GenreAnalysisResponse
	if err := json.Unmarshal([]byte(content), &analysisResp); err != nil {
		log.Printf("Failed to parse genre response, attempting raw extraction")
		genre := extractGenre(content)
		return genre, nil
	}

	log.Printf("Genre detected: %s (confidence: %.2f)", analysisResp.Genre, analysisResp.Confidence)
	return analysisResp.Genre, nil
}

func (ga *GenreAnalyzer) FindInterestingSegments(ctx context.Context, transcript, genre string) ([]models.VideoSegment, error) {
	log.Printf("Finding interesting segments for genre: %s", genre)

	prompt := ga.buildSegmentPrompt(transcript, genre)

	resp, err := ga.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.5,
		MaxTokens:   2000,
	})

	if err != nil {
		return nil, fmt.Errorf("GPT API error: %v", err)
	}

	content := resp.Choices[0].Message.Content
	var segmentResp SegmentFinderResponse
	if err := json.Unmarshal([]byte(content), &segmentResp); err != nil {
		log.Printf("Failed to parse segments response: %v", err)
		return nil, err
	}

	segments := make([]models.VideoSegment, 0)
	for _, seg := range segmentResp.Segments {
		segment := models.VideoSegment{
			StartTime:      seg.StartTime,
			EndTime:        seg.EndTime,
			Duration:       seg.EndTime - seg.StartTime,
			Reason:         seg.Reason,
			RelevanceScore: seg.RelevanceScore,
		}
		segments = append(segments, segment)
	}

	log.Printf("Found %d interesting segments", len(segments))
	return segments, nil
}

func (ga *GenreAnalyzer) buildSegmentPrompt(transcript, genre string) string {
	var instructions string

	switch strings.ToLower(genre) {
	case "comedy":
		instructions = `Find the FUNNIEST and most HILARIOUS moments. Look for:
- Punchlines and jokes
- Funny observations or reactions
- Humorous situations
- Unexpected funny moments`

	case "educational":
		instructions = `Find the MOST IMPORTANT and EDUCATIONAL moments. Look for:
- Key teachings or lessons
- Important explanations
- Practical tips
- Valuable information`

	case "sports":
		instructions = `Find the MOST EXCITING moments. Look for:
- Goals, scores, or wins
- Dramatic plays
- Unexpected moments
- Peak excitement`

	case "gaming":
		instructions = `Find the MOST EPIC moments. Look for:
- Epic plays or wins
- Funny fails
- Shocking moments
- Peak action`

	case "music":
		instructions = `Find the BEST MUSICAL moments. Look for:
- Best vocals or performances
- Most catchy parts
- Instrumental highlights
- Peak musical moments`

	default:
		instructions = `Find the MOST INTERESTING and ENGAGING moments. Look for:
- Turning points
- High emotions
- Unexpected twists
- Most entertaining parts`
	}

	prompt := fmt.Sprintf(`You are a video editor analyzing a transcript to find the best moments to clip.

Genre: %s

%s

TRANSCRIPT (with approximate timestamps):
%s

Analyze the transcript and find 3-5 of the BEST segments to clip (each 10-60 seconds long).

Respond in JSON format:
{
  "segments": [
    {
      "start_time": 0.0,
      "end_time": 30.0,
      "reason": "why this segment is interesting",
      "relevance_score": 0.85,
      "keyword": "main topic"
    }
  ]
}`, genre, instructions, truncate(transcript, 3000))

	return prompt
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func extractGenre(content string) string {
	genres := []string{"Comedy", "Educational", "Entertainment", "Sports", "Music", "News", "Gaming", "Vlog", "Tutorial", "Documentary", "Drama", "Animation"}
	for _, genre := range genres {
		if strings.Contains(content, genre) {
			return genre
		}
	}
	return "Entertainment"
}
