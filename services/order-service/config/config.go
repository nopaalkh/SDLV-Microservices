package config

import (
	"log"

	"github.com/spf13/viper"
)

// AppConfig holds the configuration for the Order Service
type AppConfig struct {
	DBHost               string `mapstructure:"DB_HOST"`
	DBPort               string `mapstructure:"DB_PORT"`
	DBUser               string `mapstructure:"DB_USER"`
	DBPassword           string `mapstructure:"DB_PASSWORD"`
	DBName               string `mapstructure:"DB_NAME"`
	JWTSecret            string `mapstructure:"JWT_SECRET"`
	Port                 string `mapstructure:"PORT"`
	IdentityServiceURL   string `mapstructure:"IDENTITY_SERVICE_URL"`
	CatalogServiceURL    string `mapstructure:"CATALOG_SERVICE_URL"`
	SMTPHost             string `mapstructure:"SMTP_HOST"`
	SMTPPort             string `mapstructure:"SMTP_PORT"`
	SMTPFrom             string `mapstructure:"SMTP_FROM"`
	MidtransServerKey    string `mapstructure:"MIDTRANS_SERVER_KEY"`
	MidtransClientKey    string `mapstructure:"MIDTRANS_CLIENT_KEY"`
	MidtransIsProduction bool   `mapstructure:"MIDTRANS_IS_PRODUCTION"`
	MidtransAPIBaseURL   string `mapstructure:"MIDTRANS_API_BASE_URL"`
	InternalAPIKey       string `mapstructure:"INTERNAL_API_KEY"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*AppConfig, error) {
	// Read .env file if available (for local development)
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	_ = viper.ReadInConfig()

	viper.AutomaticEnv()

	// Bind environment variables explicitly (required for Unmarshal)
	for _, key := range []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
		"JWT_SECRET", "PORT", "IDENTITY_SERVICE_URL", "CATALOG_SERVICE_URL",
		"SMTP_HOST", "SMTP_PORT", "SMTP_FROM",
		"MIDTRANS_SERVER_KEY", "MIDTRANS_CLIENT_KEY", "MIDTRANS_IS_PRODUCTION", "MIDTRANS_API_BASE_URL",
		"INTERNAL_API_KEY",
	} {
		_ = viper.BindEnv(key)
	}

	// Set default port
	viper.SetDefault("PORT", "3003")

	// Set default service URLs
	viper.SetDefault("IDENTITY_SERVICE_URL", "http://identity-service:3001")
	viper.SetDefault("CATALOG_SERVICE_URL", "http://catalog-service:3002")

	// Set default SMTP settings for Mailhog
	viper.SetDefault("SMTP_HOST", "mailhog")
	viper.SetDefault("SMTP_PORT", "1025")
	viper.SetDefault("SMTP_FROM", "no-reply@digitalassetmarketplace.com")

	// Set default Midtrans settings
	viper.SetDefault("MIDTRANS_IS_PRODUCTION", false)
	viper.SetDefault("MIDTRANS_API_BASE_URL", "https://api.sandbox.midtrans.com")

	// Set default Internal API Key
	viper.SetDefault("INTERNAL_API_KEY", "internal-service-key-change-me")

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
	log.Printf("Order Service Configuration:")
	log.Printf("Port: %s", c.Port)
	log.Printf("DB Host: %s", c.DBHost)
	log.Printf("DB Port: %s", c.DBPort)
	log.Printf("DB Name: %s", c.DBName)
	log.Printf("Identity Service URL: %s", c.IdentityServiceURL)
	log.Printf("Catalog Service URL: %s", c.CatalogServiceURL)
	log.Printf("SMTP Host: %s", c.SMTPHost)
	log.Printf("SMTP Port: %s", c.SMTPPort)
	log.Printf("SMTP From: %s", c.SMTPFrom)
	log.Printf("Midtrans Server Key: [HIDDEN]")
	log.Printf("Midtrans Is Production: %t", c.MidtransIsProduction)
	log.Printf("Midtrans API Base URL: %s", c.MidtransAPIBaseURL)
	log.Printf("JWT Secret: [HIDDEN]")
}
