package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	ServerPort     string
	Environment    string
	LogLevel       string
	AllowedOrigins string

	// Database (Neon PostgreSQL)
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret          string
	JWTRefreshSecret   string
	JWTExpiry          time.Duration
	RefreshTokenExpiry time.Duration

	// Google OAuth2
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// AI Providers
	GeminiAPIKey string

	// OCR
	TesseractPath string

	// Voice / Whisper
	WhisperModelPath string

	// Email (SMTP)
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	// Firebase Cloud Messaging
	FirebaseProjectID   string
	FirebaseClientEmail string
	FirebasePrivateKey  string

	// Cloudinary Storage
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string

	// Storage provider: "local" | "cloudinary"
	StorageProvider string
	LocalStoragePath string

	// Monitoring
	PrometheusEnabled bool
}

// Load reads config from .env file (if present) then environment variables.
func Load() (*Config, error) {
	// Non-fatal – .env may not exist in production
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		Environment:    getEnv("APP_ENV", "development"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),

		DatabaseURL: getEnv("DATABASE_URL", ""),

		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379/0"),

		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTRefreshSecret:   getEnv("JWT_REFRESH_SECRET", ""),
		JWTExpiry:          parseDuration("JWT_EXPIRY", 3600),
		RefreshTokenExpiry: parseDuration("REFRESH_TOKEN_EXPIRY", 604800),

		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),

		GeminiAPIKey:     getEnv("GEMINI_API_KEY", ""),
		TesseractPath:    getEnv("TESSERACT_PATH", "/usr/bin/tesseract"),
		WhisperModelPath: getEnv("WHISPER_MODEL_PATH", "./models/whisper"),

		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),

		FirebaseProjectID:   getEnv("FIREBASE_PROJECT_ID", ""),
		FirebaseClientEmail: getEnv("FIREBASE_CLIENT_EMAIL", ""),
		FirebasePrivateKey:  getEnv("FIREBASE_PRIVATE_KEY", ""),

		CloudinaryCloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:    getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret: getEnv("CLOUDINARY_API_SECRET", ""),

		StorageProvider:  getEnv("STORAGE_PROVIDER", "local"),
		LocalStoragePath: getEnv("LOCAL_STORAGE_PATH", "./uploads"),

		PrometheusEnabled: getBool("PROMETHEUS_ENABLED", true),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.JWTRefreshSecret == "" {
		return fmt.Errorf("JWT_REFRESH_SECRET is required")
	}
	return nil
}

// IsDevelopment returns true when running in development mode.
func (c *Config) IsDevelopment() bool { return c.Environment == "development" }

// IsProduction returns true when running in production mode.
func (c *Config) IsProduction() bool { return c.Environment == "production" }

// ─── helpers ──────────────────────────────────────────────────────────────────

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v == "true" || v == "1" || v == "yes"
}

func parseDuration(key string, fallbackSeconds int) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return time.Duration(fallbackSeconds) * time.Second
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return time.Duration(fallbackSeconds) * time.Second
	}
	return time.Duration(n) * time.Second
}
