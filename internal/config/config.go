package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// General
	GoEnv      string
	Port       string
	APIVersion string

	// Firebase
	FirebaseProjectID       string
	FirebaseCredentialsPath string
	FirebaseStorageBucket   string

	// Redis (Upstash)
	RedisURL      string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// JWT
	JWTSecret           string
	JWTRefreshSecret    string
	JWTExpiresIn        string
	JWTRefreshExpiresIn string

	// Mercado Pago
	MPAccessToken   string
	MPPublicKey     string
	MPWebhookSecret string

	// SendGrid
	SendGridAPIKey   string
	SendGridFromEmail string
	SendGridFromName  string

	// Cloudinary
	CloudinaryCloudName   string
	CloudinaryAPIKey      string
	CloudinaryAPISecret   string
	CloudinaryUploadPreset string

	// WhatsApp
	WhatsAppPhoneNumber        string
	WhatsAppBusinessAPIToken   string

	// CORS
	CORSAllowedOrigins string

	// Rate Limiting
	RateLimitRequests int
	RateLimitDuration string

	// Frontend
	FrontendURL string

	// Render
	RenderExternalURL string
}

func LoadConfig() (*Config, error) {
	// Load .env file if exists (for local development)
	_ = godotenv.Load()

	cfg := &Config{
		// General
		GoEnv:      getEnv("GO_ENV", "development"),
		Port:       getEnv("PORT", "8080"),
		APIVersion: getEnv("API_VERSION", "v1"),

		// Firebase
		FirebaseProjectID:       getEnv("FIREBASE_PROJECT_ID", ""),
		FirebaseCredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", "./firebase-credentials.json"),
		FirebaseStorageBucket:   getEnv("FIREBASE_STORAGE_BUCKET", ""),

		// Redis
		RedisURL:      getEnv("REDIS_URL", ""),
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// JWT
		JWTSecret:           getEnv("JWT_SECRET", ""),
		JWTRefreshSecret:    getEnv("JWT_REFRESH_SECRET", ""),
		JWTExpiresIn:        getEnv("JWT_EXPIRES_IN", "15m"),
		JWTRefreshExpiresIn: getEnv("JWT_REFRESH_EXPIRES_IN", "168h"),

		// Mercado Pago
		MPAccessToken:   getEnv("MP_ACCESS_TOKEN", ""),
		MPPublicKey:     getEnv("MP_PUBLIC_KEY", ""),
		MPWebhookSecret: getEnv("MP_WEBHOOK_SECRET", ""),

		// SendGrid
		SendGridAPIKey:    getEnv("SENDGRID_API_KEY", ""),
		SendGridFromEmail: getEnv("SENDGRID_FROM_EMAIL", "pedidos@cheoscafe.com"),
		SendGridFromName:  getEnv("SENDGRID_FROM_NAME", "Cheos Café"),

		// Cloudinary
		CloudinaryCloudName:    getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:       getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret:    getEnv("CLOUDINARY_API_SECRET", ""),
		CloudinaryUploadPreset: getEnv("CLOUDINARY_UPLOAD_PRESET", ""),

		// WhatsApp
		WhatsAppPhoneNumber:      getEnv("WHATSAPP_PHONE_NUMBER", "573001234567"),
		WhatsAppBusinessAPIToken: getEnv("WHATSAPP_BUSINESS_API_TOKEN", ""),

		// CORS
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000"),

		// Rate Limiting
		RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitDuration: getEnv("RATE_LIMIT_DURATION", "15m"),

		// Frontend
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),

		// Render
		RenderExternalURL: getEnv("RENDER_EXTERNAL_URL", ""),
	}

	// Validate critical configurations
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.FirebaseProjectID == "" && c.GoEnv == "production" {
		return fmt.Errorf("FIREBASE_PROJECT_ID is required in production")
	}

	if c.GoEnv == "production" {
		if c.JWTSecret == "" {
			return fmt.Errorf("JWT_SECRET is required in production")
		}
		if c.JWTRefreshSecret == "" {
			return fmt.Errorf("JWT_REFRESH_SECRET is required in production")
		}
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func (c *Config) IsDevelopment() bool {
	return c.GoEnv == "development"
}

func (c *Config) IsProduction() bool {
	return c.GoEnv == "production"
}
