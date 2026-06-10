package config

import (
	"log"
	"github.com/spf13/viper"
)

// AppConfig holds the configuration for the Identity Service
type AppConfig struct {
	DBHost         string `mapstructure:"DB_HOST"`
	DBPort         string `mapstructure:"DB_PORT"`
	DBUser         string `mapstructure:"DB_USER"`
	DBPassword     string `mapstructure:"DB_PASSWORD"`
	DBName         string `mapstructure:"DB_NAME"`
	JWTSecret      string `mapstructure:"JWT_SECRET"`
	Port           string `mapstructure:"PORT"`
	AdminEmail     string `mapstructure:"ADMIN_EMAIL"`
	AdminPassword  string `mapstructure:"ADMIN_PASSWORD"`
	InternalAPIKey string `mapstructure:"INTERNAL_API_KEY"`
	SMTPHost       string `mapstructure:"SMTP_HOST"`
	SMTPPort       string `mapstructure:"SMTP_PORT"`
	FrontendURL    string `mapstructure:"FRONTEND_URL"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*AppConfig, error) {
	// Read .env file if available (for local development)
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	_ = viper.ReadInConfig()

	viper.AutomaticEnv()

	// Bind environment variables explicitly (required for Unmarshal)
	for _, key := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "JWT_SECRET", "PORT", "ADMIN_EMAIL", "ADMIN_PASSWORD", "INTERNAL_API_KEY", "SMTP_HOST", "SMTP_PORT", "FRONTEND_URL"} {
		_ = viper.BindEnv(key)
	}

	// Set default port
	viper.SetDefault("PORT", "3001")
	viper.SetDefault("ADMIN_EMAIL", "admin@marketplace.com")
	viper.SetDefault("ADMIN_PASSWORD", "admin123456")
	viper.SetDefault("INTERNAL_API_KEY", "internal-service-key-change-me")
	viper.SetDefault("SMTP_HOST", "mailhog")
	viper.SetDefault("SMTP_PORT", "1025")
	viper.SetDefault("FRONTEND_URL", "http://localhost:5173")

	var config AppConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetDBConnectionString returns the PostgreSQL connection string
func (c *AppConfig) GetDBConnectionString() string {
	return "host=" + c.DBHost + " port=" + c.DBPort + " user=" + c.DBUser + " password=" + c.DBPassword + " dbname=" + c.DBName + " sslmode=disable"
}

// LogConfig logs the current configuration (without sensitive data)
func (c *AppConfig) LogConfig() {
	log.Printf("Identity Service Configuration:")
	log.Printf("Port: %s", c.Port)
	log.Printf("DB Host: %s", c.DBHost)
	log.Printf("DB Port: %s", c.DBPort)
	log.Printf("DB Name: %s", c.DBName)
	log.Printf("JWT Secret: [HIDDEN]")
}