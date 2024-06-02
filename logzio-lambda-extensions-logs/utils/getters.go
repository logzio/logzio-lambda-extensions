package utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

var logger = log.WithFields(log.Fields{"agent": "logzioLogsExtension"})

const (
	envLogzioLogsToken         = "LOGZIO_LOGS_TOKEN"
	envLogzioListener          = "LOGZIO_LISTENER"
	envEnablePlatformLogs      = "ENABLE_PLATFORM_LOGS"
	envExtensionLogLevel       = "LOGS_EXT_LOG_LEVEL"
	envGrokPatterns            = "GROK_PATTERNS"
	envLogsFormat              = "LOGS_FORMAT"
	envCustomFields            = "CUSTOM_FIELDS"
	envFlattenNestedMessage    = "FLATTEN_NESTED_MESSAGE"
	envAwsLambdaFunctionName   = "AWS_LAMBDA_FUNCTION_NAME" // Reserved AWS env var
	envAwsRegion               = "AWS_REGION"               //Reserved AWS env var
	LogLevelDebug              = "debug"
	LogLevelInfo               = "info"
	LogLevelWarn               = "warn"
	LogLevelError              = "error"
	LogLevelFatal              = "fatal"
	LogLevelPanic              = "panic"
	defaultEnablePlatformLogs  = false
	defaultEnableExtensionLogs = false
	defaultExtensionLogLevel   = LogLevelInfo
)

func GetToken() (string, error) {
	token := os.Getenv(envLogzioLogsToken)
	if len(token) == 0 {
		return "", fmt.Errorf("%s must be set", envLogzioLogsToken)
	}

	return token, nil
}

func GetListener() (string, error) {
	token := os.Getenv(envLogzioListener)
	if len(token) == 0 {
		return "", fmt.Errorf("%s must be set", envLogzioListener)
	}

	return token, nil
}

func GetEnablePlatformLogs() bool {
	enableStr := os.Getenv(envEnablePlatformLogs)
	if len(enableStr) == 0 {
		return defaultEnablePlatformLogs
	}

	enable, err := strconv.ParseBool(enableStr)
	if err != nil {
		logger.Warningf("Could not parse env var %s, reverting to default value (%t)", envEnablePlatformLogs, defaultEnablePlatformLogs)
		return defaultEnablePlatformLogs
	}

	return enable
}

func GetExtensionLogLevel() string {
	validLogLevels := []string{LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError, LogLevelFatal, LogLevelPanic}
	logLevel := os.Getenv(envExtensionLogLevel)

	for _, validLogLevel := range validLogLevels {
		if validLogLevel == logLevel {
			return validLogLevel
		}
	}

	logger.Infof("Reverting to default log level: %s", defaultExtensionLogLevel)
	return defaultExtensionLogLevel
}

func GetGrokPatterns() string {
	return os.Getenv(envGrokPatterns)
}

func GetLogsFormat() string {
	return os.Getenv(envLogsFormat)
}

func GetAwsLambdaFunctionName() string {
	return os.Getenv(envAwsLambdaFunctionName)
}

func GetAwsRegion() string {
	return os.Getenv(envAwsRegion)
}

func GetCustomFields() map[string]string {
	keyIndex := 0
	valueIndex := 1
	fieldsStr := os.Getenv(envCustomFields)
	customFields := make(map[string]string, 0)
	if len(fieldsStr) == 0 {
		return customFields
	}

	pairs := strings.Split(fieldsStr, ",")
	for _, pair := range pairs {
		keyValue := strings.Split(pair, "=")
		customFields[keyValue[keyIndex]] = keyValue[valueIndex]
	}

	logger.Debugf("detected %d custom fields", len(customFields))
	return customFields
}

func GetFlattenNestedMessage() bool {
	return strings.EqualFold("true", os.Getenv(envFlattenNestedMessage))
}
