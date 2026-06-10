package config

import (
	"log"
	"github.com/spf13/viper"
)

// AppConfig holds the configuration for the Catalog Service
type AppConfig struct {
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	Port            string `mapstructure:"PORT"`
	JWTSecret       string `mapstructure:"JWT_SECRET"`
	OrderServiceURL    string `mapstructure:"ORDER_SERVICE_URL"`
	IdentityServiceURL string `mapstructure:"IDENTITY_SERVICE_URL"`
	InternalAPIKey     string `mapstructure:"INTERNAL_API_KEY"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*AppConfig, error) {
	// Read .env file if available (for local development)
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	_ = viper.ReadInConfig()

	viper.AutomaticEnv()

	// Bind environment variables explicitly (required for Unmarshal)
	for _, key := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "JWT_SECRET", "PORT", "ORDER_SERVICE_URL", "IDENTITY_SERVICE_URL", "INTERNAL_API_KEY"} {
		_ = viper.BindEnv(key)
	}

	// Set default port
	viper.SetDefault("PORT", "3002")

	var config AppConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if config.JWTSecret == "" {
		log.Println("WARNING: JWT_SECRET is not set, using default key")
		config.JWTSecret = "default-secret-key-change-me"
	}
	if config.OrderServiceURL == "" {
		config.OrderServiceURL = "http://order-service:8081"
	}
	if config.IdentityServiceURL == "" {
		config.IdentityServiceURL = "http://identity-service:3001"
	}
	if config.InternalAPIKey == "" {
		config.InternalAPIKey = "internal-service-key-change-me"
	}

	return &config, nil
}

// GetDBConnectionString returns the PostgreSQL connection string
func (c *AppConfig) GetDBConnectionString() string {
	return "host=" + c.DBHost + " port=" + c.DBPort + " user=" + c.DBUser + " password=" + c.DBPassword + " dbname=" + c.DBName + " sslmode=disable"
}

// LogConfig logs the current configuration (without sensitive data)
func (c *AppConfig) LogConfig() {
	log.Printf("Catalog Service Configuration:")
	log.Printf("Port: %s", c.Port)
	log.Printf("DB Host: %s", c.DBHost)
	log.Printf("DB Port: %s", c.DBPort)
	log.Printf("DB Name: %s", c.DBName)
	log.Printf("JWT Secret: [HIDDEN]")
}