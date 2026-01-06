package config

import (
	"os"
	"rest_api_poc/internal/shared/logger"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type WebServerConfig struct {
	Env           string
	Port          string
	CORSOrigins   []string
	EnableSwagger bool
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
}

type DBConfig struct {
	ConnectionString string
	DBRetryCount     int
}

type CacheConfig struct {
	Enable   bool
	Address  string
	Password string
	DB       int
	TTL      time.Duration
}

type AuthConfig struct {
	JWTSecret                string
	JWTIssuer                string
	Audience                 []string
	AccessTokenLifetime      time.Duration
	RefreshTokenLifetime     time.Duration
	StaySignedInLifetime     time.Duration
	PasswordResetOTPLifetime time.Duration
}

type Config struct {
	WebServer WebServerConfig
	DB        DBConfig
	Cache     CacheConfig
	Auth      AuthConfig
}

// -------------------------
// Helper functions
// -------------------------
func getEnv(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func mustGetEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		logger.Fatal("Missing environment variable %s", key)
	}
	return val
}

func getEnvAsBool(key string, defaultVal bool) bool {
	valStr, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		logger.Fatal("Invalid bool value for %s, %v", key, err)
	}
	return val
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valStr, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	val, err := time.ParseDuration(valStr)
	if err != nil {
		logger.Fatal("Invalid duration for %s, %v", key, err)
	}
	return val
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		logger.Fatal("Invalid int for env %s: %v", key, err)
	}
	return val
}

// -------------------------
// Subsystem loaders
// -------------------------
func loadWebServerConfig() WebServerConfig {
	return WebServerConfig{
		Env:           getEnv("ENV", "dev"),
		Port:          getEnv("WEB_PORT", "8080"),
		CORSOrigins:   strings.Split(getEnv("CORS_ORIGINS", "*"), ","),
		EnableSwagger: getEnvAsBool("ENABLE_SWAGGER", true),
		ReadTimeout:   getEnvAsDuration("READ_TIMEOUT", 5*time.Second),
		WriteTimeout:  getEnvAsDuration("WRITE_TIMEOUT", 5*time.Second),
	}
}

func loadDBConfig() DBConfig {
	return DBConfig{
		ConnectionString: mustGetEnv("DB_CONNECTION_STRING"),
		DBRetryCount:     getEnvAsInt("DB_RETRY_COUNT", 3),
	}
}

func loadCacheConfig() CacheConfig {
	cfg := CacheConfig{
		Enable: getEnvAsBool("ENABLE_CACHE", false),
	}

	if cfg.Enable {
		cfg.Address = getEnv("REDIS_ADDRESS", "localhost:6379")
		cfg.Password = os.Getenv("REDIS_PASSWORD")
		cfg.DB = getEnvAsInt("REDIS_DB", 0)
		cfg.TTL = getEnvAsDuration("REDIS_TTL", time.Hour)
	}

	return cfg
}

func loadAuthConfig() AuthConfig {
	cfg := AuthConfig{
		JWTSecret:                mustGetEnv("JWT_SECRET"),
		JWTIssuer:                getEnv("JWT_ISSUER", "go-rest-api-poc"),
		AccessTokenLifetime:      getEnvAsDuration("ACCESS_TOKEN_LIFETIME", 15*time.Minute),
		RefreshTokenLifetime:     getEnvAsDuration("REFRESH_TOKEN_LIFETIME", 168*time.Hour),      // 7 days
		StaySignedInLifetime:     getEnvAsDuration("STAY_SIGNED_IN_LIFETIME", 720*time.Hour),     // 30 days
		PasswordResetOTPLifetime: getEnvAsDuration("PASSWORD_RESET_OTP_LIFETIME", 15*time.Minute),
	}

	if aud, ok := os.LookupEnv("JWT_AUDIENCE"); ok {
		cfg.Audience = strings.Split(aud, ",")
	} else {
		cfg.Audience = []string{"go-rest-api-poc"}
	}

	return cfg
}

func LoadConfig() *Config {
	logger.Info("loading config...")

	// load env file from local & check for error
	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found, using system environment")
	}

	config := &Config{}

	config.WebServer = loadWebServerConfig()
	config.DB = loadDBConfig()
	config.Cache = loadCacheConfig()
	config.Auth = loadAuthConfig()

	logger.Info("config is successfully loaded!!!")
	return config
}
