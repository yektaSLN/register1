package config

import (
	"os"

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
	//adding redis
	RedisAddr string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

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

		//access token will expire after 10 minutes
		JWTExpiration: 2,

		//adding redis
		RedisAddr: getEnv("REDIS_ADDR", ""),
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}
