package tmdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client represents a TMDB API client
type Client struct {
	BaseURL string
	APIKey  string
	client  *http.Client
}

// Movie represents a movie from TMDB API
type Movie struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	PosterPath       string  `json:"poster_path"`
	BackdropPath     string  `json:"backdrop_path"`
	ReleaseDate      string  `json:"release_date"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	Popularity       float64 `json:"popularity"`
	Adult            bool    `json:"adult"`
	Video            bool    `json:"video"`
	OriginalLanguage string  `json:"original_language"`
	GenreIDs         []int   `json:"genre_ids"`
}

// MovieResponse represents the response structure from TMDB search/list endpoints
type MovieResponse struct {
	Page         int     `json:"page"`
	Results      []Movie `json:"results"`
	TotalPages   int     `json:"total_pages"`
	TotalResults int     `json:"total_results"`
}

// NewClient creates a new TMDB client with a configured HTTP client
func NewClient(apiKey string) *Client {
	return &Client{
		BaseURL: "https://api.themoviedb.org/3",
		APIKey:  apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SearchMovie searches for movies by query string
func (c *Client) SearchMovie(query string) (*MovieResponse, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	endpoint := fmt.Sprintf("%s/search/movie", c.BaseURL)

	params := url.Values{}
	params.Add("api_key", c.APIKey)
	params.Add("query", query)
	params.Add("include_adult", "false")

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var movieResp MovieResponse
	if err := json.NewDecoder(resp.Body).Decode(&movieResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &movieResp, nil
}

// GetNowPlaying retrieves currently playing movies in theaters
func (c *Client) GetNowPlaying() (*MovieResponse, error) {
	endpoint := fmt.Sprintf("%s/movie/now_playing", c.BaseURL)

	params := url.Values{}
	params.Add("api_key", c.APIKey)

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var movieResp MovieResponse
	if err := json.NewDecoder(resp.Body).Decode(&movieResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &movieResp, nil
}
