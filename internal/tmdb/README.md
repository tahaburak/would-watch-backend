# TMDB Client

A Go client for interacting with The Movie Database (TMDB) API.

## Features

- Search for movies by query string
- Get currently playing movies in theaters
- Type-safe response structures
- Configurable HTTP timeout (10 seconds)
- Comprehensive error handling

## Usage

```go
import "github.com/burak/would_watch/backend/internal/tmdb"

// Create a new client
client := tmdb.NewClient("your-tmdb-api-key")

// Search for movies
results, err := client.SearchMovie("The Matrix")
if err != nil {
    log.Fatal(err)
}

for _, movie := range results.Results {
    fmt.Printf("Found: %s (%s)\n", movie.Title, movie.ReleaseDate)
}

// Get now playing movies
nowPlaying, err := client.GetNowPlaying()
if err != nil {
    log.Fatal(err)
}

for _, movie := range nowPlaying.Results {
    fmt.Printf("Now Playing: %s (Rating: %.1f)\n", movie.Title, movie.VoteAverage)
}
```

## Response Structure

The `MovieResponse` struct contains:
- `Page`: Current page number
- `Results`: Array of `Movie` objects
- `TotalPages`: Total number of pages
- `TotalResults`: Total number of results

Each `Movie` object includes:
- `ID`: TMDB movie ID
- `Title`: Movie title
- `Overview`: Movie description
- `PosterPath`: Path to poster image
- `ReleaseDate`: Release date
- `VoteAverage`: Average rating
- And more fields...

## Testing

Run tests with:
```bash
go test ./internal/tmdb/... -v
```
