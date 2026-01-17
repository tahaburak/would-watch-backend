package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client represents an OpenAI API client
type Client struct {
	APIKey  string
	BaseURL string
	client  *http.Client
}

// ChatRequest represents the OpenAI chat completion request
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

// ChatMessage represents a message in the chat
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents the OpenAI chat completion response
type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

// NewClient creates a new OpenAI client
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: "https://api.openai.com/v1",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetRecommendations gets movie recommendations based on liked movies
// Returns TMDB IDs of recommended movies
func (c *Client) GetRecommendations(likedMovies []string) ([]int, error) {
	if len(likedMovies) == 0 {
		return nil, fmt.Errorf("no liked movies provided")
	}

	// Create the prompt
	movieList := strings.Join(likedMovies, ", ")
	prompt := fmt.Sprintf(`You are a movie expert. Given these movies that users liked: [%s], recommend 5 distinct movies that they would enjoy. Return ONLY a JSON array of TMDB IDs as integers, nothing else. Example format: [123, 456, 789, 101, 202]`, movieList)

	// Prepare request
	reqBody := ChatRequest{
		Model: "gpt-4o-mini",
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from API")
	}

	// Extract TMDB IDs from the response
	content := chatResp.Choices[0].Message.Content
	content = strings.TrimSpace(content)

	// Parse JSON array of IDs
	var tmdbIDs []int
	if err := json.Unmarshal([]byte(content), &tmdbIDs); err != nil {
		return nil, fmt.Errorf("failed to parse TMDB IDs from response: %w (content: %s)", err, content)
	}

	if len(tmdbIDs) == 0 {
		return nil, fmt.Errorf("no TMDB IDs returned")
	}

	return tmdbIDs, nil
}
