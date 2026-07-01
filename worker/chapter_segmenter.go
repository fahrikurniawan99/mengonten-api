package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type ChapterSegmenter struct {
	APIKey string
	Client *openai.Client
}

type SegmentChapterRequest struct {
	Transcript string `json:"transcript"`
}

type ChapterResponse struct {
	Chapters []ChapterData `json:"chapters"`
}

type ChapterData struct {
	Index       int      `json:"index"`
	StartTime   float64  `json:"start_time"`
	EndTime     float64  `json:"end_time"`
	Duration    float64  `json:"duration"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

func NewChapterSegmenter(apiKey string) *ChapterSegmenter {
	baseURL := os.Getenv("GROQ_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.groq.com/openai/v1"
	}

	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = baseURL

	return &ChapterSegmenter{
		APIKey: apiKey,
		Client: openai.NewClientWithConfig(cfg),
	}
}

func (cs *ChapterSegmenter) SegmentChapters(ctx context.Context, transcript string) ([]ChapterData, error) {
	log.Printf("[ChapterSegmenter] Starting segmentation for %d chars transcript", len(transcript))

	prompt := cs.buildChapterPrompt(transcript)

	resp, err := cs.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.3,
		MaxTokens:   2000,
	})

	if err != nil {
		return nil, fmt.Errorf("groq API error: %v", err)
	}

	content := resp.Choices[0].Message.Content
	chapters, err := cs.parseChaptersResponse(content)
	if err != nil {
		log.Printf("[ChapterSegmenter] Parse error: %v, raw: %s", err, content)
		return nil, err
	}

	log.Printf("[ChapterSegmenter] Found %d chapters", len(chapters))
	return chapters, nil
}

func (cs *ChapterSegmenter) buildChapterPrompt(transcript string) string {
	truncated := transcript
	if len(transcript) > 5000 {
		truncated = transcript[:5000] + "..."
	}

	prompt := fmt.Sprintf(`You are a video content analyzer. Analyze this video transcript and identify distinct chapters/topics.

INSTRUCTIONS:
1. Detect natural topic shifts and content boundaries
2. Each chapter should represent a cohesive topic or section
3. Provide timestamps (in seconds) for start and end of each chapter
4. Generate a concise title and description for each chapter
5. Extract 2-3 key keywords per chapter

TRANSCRIPT (with approximate timestamps):
%s

Respond ONLY in valid JSON format (no markdown, no explanation):
{
  "chapters": [
    {
      "index": 1,
      "start_time": 0.0,
      "end_time": 45.5,
      "duration": 45.5,
      "title": "Chapter Title",
      "description": "Brief description of chapter content",
      "keywords": ["keyword1", "keyword2", "keyword3"]
    }
  ]
}`, truncated)

	return prompt
}

func (cs *ChapterSegmenter) parseChaptersResponse(content string) ([]ChapterData, error) {
	content = strings.TrimSpace(content)

	// Remove markdown code blocks if present
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	var resp ChapterResponse
	if err := json.Unmarshal([]byte(content), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse chapters JSON: %v", err)
	}

	if len(resp.Chapters) == 0 {
		return nil, fmt.Errorf("no chapters found in response")
	}

	return resp.Chapters, nil
}
