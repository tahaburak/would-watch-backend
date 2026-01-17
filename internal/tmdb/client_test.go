package tmdb

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"
	client := NewClient(apiKey)

	if client.APIKey != apiKey {
		t.Errorf("expected APIKey %s, got %s", apiKey, client.APIKey)
	}

	if client.BaseURL != "https://api.themoviedb.org/3" {
		t.Errorf("expected BaseURL https://api.themoviedb.org/3, got %s", client.BaseURL)
	}

	if client.client == nil {
		t.Error("expected http client to be initialized")
	}
}

func TestSearchMovie_EmptyQuery(t *testing.T) {
	client := NewClient("test-key")
	_, err := client.SearchMovie("")

	if err == nil {
		t.Error("expected error for empty query, got nil")
	}
}

func TestSearchMovie_Success(t *testing.T) {
	mockResponse := `{
		"page": 1,
		"results": [
			{
				"id": 550,
				"title": "Fight Club",
				"original_title": "Fight Club",
				"overview": "A ticking-time-bomb insomniac...",
				"poster_path": "/path.jpg",
				"backdrop_path": "/backdrop.jpg",
				"release_date": "1999-10-15",
				"vote_average": 8.4,
				"vote_count": 26000,
				"popularity": 61.416,
				"adult": false,
				"video": false,
				"original_language": "en",
				"genre_ids": [18, 53, 35]
			}
		],
		"total_pages": 1,
		"total_results": 1
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("query") == "" {
			t.Error("expected query parameter")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := NewClient("test-key")
	client.BaseURL = server.URL

	resp, err := client.SearchMovie("Fight Club")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Page != 1 {
		t.Errorf("expected page 1, got %d", resp.Page)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}

	movie := resp.Results[0]
	if movie.Title != "Fight Club" {
		t.Errorf("expected title 'Fight Club', got %s", movie.Title)
	}

	if movie.ID != 550 {
		t.Errorf("expected id 550, got %d", movie.ID)
	}
}

func TestGetNowPlaying_Success(t *testing.T) {
	mockResponse := `{
		"page": 1,
		"results": [
			{
				"id": 12345,
				"title": "Now Playing Movie",
				"original_title": "Now Playing Movie",
				"overview": "A movie currently in theaters",
				"poster_path": "/poster.jpg",
				"backdrop_path": "/backdrop.jpg",
				"release_date": "2024-01-15",
				"vote_average": 7.5,
				"vote_count": 1000,
				"popularity": 50.0,
				"adult": false,
				"video": false,
				"original_language": "en",
				"genre_ids": [28, 12]
			}
		],
		"total_pages": 5,
		"total_results": 100
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := NewClient("test-key")
	client.BaseURL = server.URL

	resp, err := client.GetNowPlaying()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Page != 1 {
		t.Errorf("expected page 1, got %d", resp.Page)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}

	movie := resp.Results[0]
	if movie.Title != "Now Playing Movie" {
		t.Errorf("expected title 'Now Playing Movie', got %s", movie.Title)
	}
}

func TestSearchMovie_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status_message":"Invalid API key"}`))
	}))
	defer server.Close()

	client := NewClient("invalid-key")
	client.BaseURL = server.URL

	_, err := client.SearchMovie("test")
	if err == nil {
		t.Error("expected error for API failure, got nil")
	}
}

func TestGetNowPlaying_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status_message":"Internal error"}`))
	}))
	defer server.Close()

	client := NewClient("test-key")
	client.BaseURL = server.URL

	_, err := client.GetNowPlaying()
	if err == nil {
		t.Error("expected error for API failure, got nil")
	}
}
