package utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

var logger = log.WithFields(log.Fields{"agent": "logzioLogsExtension"})

const (
	envLogzioLogsToken         = "LOGZIO_LOGS_TOKEN"
	envLogzioListener          = "LOGZIO_LISTENER"
	envEnablePlatformLogs      = "ENABLE_PLATFORM_LOGS"
	envEnableExtensionLogs     = "ENABLE_EXTENSION_LOGS"
	envExtensionLogLevel       = "LOGS_EXT_LOG_LEVEL"
	envGrokPatterns            = "GROK_PATTERNS"
	envLogsFormat              = "LOGS_FORMAT"
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

func GetEnableExtensionLogs() bool {
	enableStr := os.Getenv(envEnableExtensionLogs)
	if len(enableStr) == 0 {
		return defaultEnableExtensionLogs
	}

	enable, err := strconv.ParseBool(enableStr)
	if err != nil {
		logger.Warningf("Could not parse env var %s, reverting to default value (%t)", envEnableExtensionLogs, defaultEnableExtensionLogs)
		return defaultEnableExtensionLogs
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
