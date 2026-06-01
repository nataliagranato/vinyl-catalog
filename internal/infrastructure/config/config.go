package config

import "os"

type Config struct {
	AppPort string
	AppEnv  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret          string
	JWTExpirationHours string
	AdminUsername      string
	AdminPassword      string

	OTELEndpoint    string
	OTELServiceName string
	ServiceVersion  string
}

func Load() *Config {
	return &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "vinyl_catalog"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:          getEnv("JWT_SECRET", "change-me"),
		JWTExpirationHours: getEnv("JWT_EXPIRATION_HOURS", "24"),
		AdminUsername:      getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:      getEnv("ADMIN_PASSWORD", "admin"),

		OTELEndpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		OTELServiceName: getEnv("OTEL_SERVICE_NAME", "vinyl-catalog"),
		ServiceVersion:  getEnv("SERVICE_VERSION", "1.0.0"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
