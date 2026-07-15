package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
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
	SMTPUser      string
	UseSESSMTP    string
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

	optionalParams := map[string]*string{
		"SMTP_USER":    &secrets.SMTPUser,
		"USE_SES_SMTP": &secrets.UseSESSMTP,
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

	for key, dest := range optionalParams {
		if err := loadOptionalSSMParameter(ctx, ssmClient, basePath, key, dest); err != nil {
			return nil, err
		}
	}

	if strings.TrimSpace(secrets.UseSESSMTP) == "" {
		secrets.UseSESSMTP = "false"
	}

	appSecrets = secrets
	return secrets, nil
}

func loadOptionalSSMParameter(ctx context.Context, ssmClient *ssm.Client, basePath, key string, dest *string) error {
	paramName := fmt.Sprintf("%s/%s", basePath, key)
	result, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		var notFound *types.ParameterNotFound
		if errors.As(err, &notFound) {
			return nil
		}
		return fmt.Errorf("failed to get parameter %s: %w", paramName, err)
	}
	*dest = *result.Parameter.Value
	return nil
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
		SMTPUser:      getEnv("SMTP_USER", ""),
		UseSESSMTP:    getEnv("USE_SES_SMTP", "false"),
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

// Get returns the loaded secrets (auto-loads if not already loaded)
func Get() *Secrets {
	if appSecrets == nil {
		secrets, err := LoadSecrets()
		if err != nil {
			panic("Failed to load secrets: " + err.Error())
		}
		return secrets
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
	case "SMTP_USER":
		return secrets.SMTPUser
	case "USE_SES_SMTP":
		return secrets.UseSESSMTP
	case "MONTH_INTERVAL":
		return secrets.MonthInterval
	default:
		return ""
	}
}

// UseSESSMTP is true when Amazon SES SMTP credentials should be used (SMTP_USER + SES host).
// Default false keeps legacy Office365 behavior (EMAIL_FROM as SMTP username).
func UseSESSMTP() bool {
	return parseTruthy(configString("USE_SES_SMTP", "false"))
}

// SMTPAuthUsername returns the SMTP login user for the active email provider.
func SMTPAuthUsername() string {
	secrets := Get()
	if UseSESSMTP() {
		return strings.TrimSpace(secrets.SMTPUser)
	}
	return secrets.EmailFrom
}

func configString(key, defaultValue string) string {
	value := strings.TrimSpace(GetValue(key))
	if value == "" {
		return defaultValue
	}
	return value
}

func parseTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

