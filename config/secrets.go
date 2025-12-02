package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// Secrets holds all application secrets
type Secrets struct {
	SecretKey     string
	MongoURI      string
	MongoDB       string
	EmailFrom     string
	EmailPassword string
	EmailPort     string
	SMTPClient    string
	MonthInterval string
}

var appSecrets *Secrets

// LoadSecrets loads secrets from AWS Parameter Store based on environment
func LoadSecrets() (*Secrets, error) {
	if appSecrets != nil {
		return appSecrets, nil
	}

	// Get environment (default to "dev")
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "dev"
	}

	// Check if we should use local .env (for local development)
	useLocal := os.Getenv("USE_LOCAL_ENV")
	if useLocal == "true" {
		return loadFromEnv(), nil
	}

	// Load from AWS Parameter Store
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	ssmClient := ssm.NewFromConfig(cfg)
	basePath := fmt.Sprintf("/lbtechapi/%s", environment)

	secrets := &Secrets{}

	// Load each parameter
	params := map[string]*string{
		"SECRET_KEY":      &secrets.SecretKey,
		"MONGO_URI":       &secrets.MongoURI,
		"MONGO_DB":        &secrets.MongoDB,
		"EMAIL_FROM":      &secrets.EmailFrom,
		"EMAIL_PASSWORD":  &secrets.EmailPassword,
		"EMAIL_PORT":      &secrets.EmailPort,
		"SMTP_CLIENTE":    &secrets.SMTPClient,
		"MONTH_INTERVAL":  &secrets.MonthInterval,
	}

	for key, dest := range params {
		paramName := fmt.Sprintf("%s/%s", basePath, key)
		result, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
			Name:           aws.String(paramName),
			WithDecryption: aws.Bool(true),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get parameter %s: %w", paramName, err)
		}
		*dest = *result.Parameter.Value
	}

	appSecrets = secrets
	return secrets, nil
}

// loadFromEnv loads secrets from environment variables (for local development)
func loadFromEnv() *Secrets {
	return &Secrets{
		SecretKey:     getEnv("SECRET_KEY", ""),
		MongoURI:      getEnv("MONGO_URI", ""),
		MongoDB:       getEnv("MONGO_DB", ""),
		EmailFrom:     getEnv("EMAIL_FROM", ""),
		EmailPassword: getEnv("EMAIL_PASSWORD", ""),
		EmailPort:     getEnv("EMAIL_PORT", "587"),
		SMTPClient:    getEnv("SMTP_CLIENTE", ""),
		MonthInterval: getEnv("MONTH_INTERVAL", "1"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Get returns the loaded secrets
func Get() *Secrets {
	if appSecrets == nil {
		panic("Secrets not loaded. Call LoadSecrets() first.")
	}
	return appSecrets
}

// GetValue gets a specific secret value by key name
func GetValue(key string) string {
	secrets := Get()
	switch strings.ToUpper(key) {
	case "SECRET_KEY":
		return secrets.SecretKey
	case "MONGO_URI":
		return secrets.MongoURI
	case "MONGO_DB":
		return secrets.MongoDB
	case "EMAIL_FROM":
		return secrets.EmailFrom
	case "EMAIL_PASSWORD":
		return secrets.EmailPassword
	case "EMAIL_PORT":
		return secrets.EmailPort
	case "SMTP_CLIENTE":
		return secrets.SMTPClient
	case "MONTH_INTERVAL":
		return secrets.MonthInterval
	default:
		return ""
	}
}

