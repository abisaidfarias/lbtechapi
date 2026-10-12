package config

import (
	"context"
	"errors"
	"fmt"
	"log"
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

	var secrets *Secrets
	var err error

	useLocal := os.Getenv("USE_LOCAL_ENV")
	if useLocal == "true" {
		secrets = loadFromEnv()
	} else {
		secrets, err = loadSecretsFromSSM(environment)
		if err != nil {
			return nil, err
		}
	}

	// Docker on EB often has no .env file inside the container; fill missing vars from SSM.
	hydrateMissingFromSSM(environment, secrets)

	if strings.TrimSpace(secrets.UseSESSMTP) == "" {
		secrets.UseSESSMTP = "false"
	}

	appSecrets = secrets
	logPasswordEmailReadiness()
	return secrets, nil
}

func loadSecretsFromSSM(environment string) (*Secrets, error) {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	ssmClient := ssm.NewFromConfig(cfg)
	basePath := fmt.Sprintf("/lbtechapi/%s", environment)

	secrets := &Secrets{}

	params := map[string]*string{
		"SECRET_KEY":     &secrets.SecretKey,
		"MONGO_URI":      &secrets.MongoURI,
		"MONGO_DB":       &secrets.MongoDB,
		"EMAIL_FROM":     &secrets.EmailFrom,
		"EMAIL_PASSWORD": &secrets.EmailPassword,
		"EMAIL_PORT":     &secrets.EmailPort,
		"SMTP_CLIENTE":   &secrets.SMTPClient,
		"MONTH_INTERVAL": &secrets.MonthInterval,
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

	return secrets, nil
}

// hydrateMissingFromSSM backfills env vars (and Secrets fields) not present in the process
// environment — required for password reset links on Elastic Beanstalk Docker.
func hydrateMissingFromSSM(environment string, secrets *Secrets) {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		return
	}
	ssmClient := ssm.NewFromConfig(cfg)
	basePath := fmt.Sprintf("/lbtechapi/%s", environment)

	runtimeKeys := []string{
		"FRONTEND_BASE_URL",
		"PASSWORD_TOKEN_TTL_HOURS",
		"DEV_EMAIL_TO",
	}
	secretKeys := map[string]*string{
		"SECRET_KEY":     &secrets.SecretKey,
		"MONGO_URI":      &secrets.MongoURI,
		"MONGO_DB":       &secrets.MongoDB,
		"EMAIL_FROM":     &secrets.EmailFrom,
		"EMAIL_PASSWORD": &secrets.EmailPassword,
		"EMAIL_PORT":     &secrets.EmailPort,
		"SMTP_CLIENTE":   &secrets.SMTPClient,
		"SMTP_USER":      &secrets.SMTPUser,
		"USE_SES_SMTP":   &secrets.UseSESSMTP,
	}

	tryHydrate := func(key string, dest *string) {
		if dest != nil && strings.TrimSpace(*dest) != "" {
			return
		}
		if strings.TrimSpace(os.Getenv(key)) != "" {
			if dest != nil {
				*dest = strings.TrimSpace(os.Getenv(key))
			}
			return
		}
		val, err := fetchSSMParameter(ctx, ssmClient, basePath, key)
		if err != nil || strings.TrimSpace(val) == "" {
			return
		}
		_ = os.Setenv(key, val)
		if dest != nil {
			*dest = val
		}
	}

	for _, key := range runtimeKeys {
		tryHydrate(key, nil)
	}
	for key, dest := range secretKeys {
		tryHydrate(key, dest)
	}
}

func fetchSSMParameter(ctx context.Context, ssmClient *ssm.Client, basePath, key string) (string, error) {
	paramName := fmt.Sprintf("%s/%s", basePath, key)
	result, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		var notFound *types.ParameterNotFound
		if errors.As(err, &notFound) {
			return "", nil
		}
		return "", err
	}
	if result.Parameter == nil || result.Parameter.Value == nil {
		return "", nil
	}
	return strings.TrimSpace(*result.Parameter.Value), nil
}

func logPasswordEmailReadiness() {
	base := strings.TrimSpace(os.Getenv("FRONTEND_BASE_URL"))
	from := strings.TrimSpace(GetValue("EMAIL_FROM"))
	if base == "" {
		log.Println("⚠️  FRONTEND_BASE_URL is empty — invite/reset password emails will not send")
	}
	if from == "" {
		log.Println("⚠️  EMAIL_FROM is empty — transactional email will not send")
	}
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

