package config

import (
	"os"
)

type Config struct {
	TMDBAPIKey      string
	OpenAIAPIKey    string
	SupabaseURL     string
	SupabaseKey     string
	SupabaseJWTSecret string
	Port            string
}

func LoadConfig() *Config {
	return &Config{
		TMDBAPIKey:        getEnv("TMDB_API_KEY", ""),
		OpenAIAPIKey:      getEnv("OPENAI_API_KEY", ""),
		SupabaseURL:       getEnv("SUPABASE_URL", ""),
		SupabaseKey:       getEnv("SUPABASE_ANON_KEY", ""),
		SupabaseJWTSecret: getEnv("SUPABASE_JWT_SECRET", ""),
		Port:              getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
