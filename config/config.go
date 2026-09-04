package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string

	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	ProductDBName string
	DBSSLMode     string

	JWTSecret     string
	JWTExpiration int

	RedisAddr string

	KafkaBrokers []string
	KafkaTopic   string

	LogFile string

	RateLimitRequests      int
	RateLimitWindowSeconds int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	rateLimitRequests, err := strconv.Atoi(
		getEnv("RATE_LIMIT_REQUESTS", "5"),
	)
	if err != nil {
		return nil, err
	}

	rateLimitWindowSeconds, err := strconv.Atoi(
		getEnv("RATE_LIMIT_WINDOW_SECONDS", "60"),
	)
	if err != nil {
		return nil, err
	}

	kafkaBrokers := strings.Split(
		getEnv("KAFKA_BROKERS", "localhost:9092"),
		",",
	)

	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),

		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", ""),
		DBName:        getEnv("DB_NAME", "login"),
		ProductDBName: getEnv("PRODUCT_DB_NAME", "products"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),

		JWTSecret: getEnv("JWT_SECRET", ""),

		JWTExpiration: 10,

		RedisAddr: getEnv("REDIS_ADDR", ""),

		KafkaBrokers: kafkaBrokers,
		KafkaTopic:   getEnv("KAFKA_TOPIC", "application-events"),

		LogFile: getEnv(
			"LOG_FILE",
			"/tmp/register1-logs/application-events.log",
		),

		RateLimitRequests:      rateLimitRequests,
		RateLimitWindowSeconds: rateLimitWindowSeconds,
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}
