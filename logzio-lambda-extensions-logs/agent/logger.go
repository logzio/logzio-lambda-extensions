package agent

import (
	"github.com/logzio/logzio-go"
	"logzio-lambda-extensions-logs/utils"

	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	maxBulkSizeBytes = 10 * 1024 * 1024 // 10 MB
)

var logger = log.WithFields(log.Fields{"agent": "logsApiAgent"})

// NewLogzioLogger returns a Logzio Sender
func NewLogzioLogger() (*logzio.LogzioSender, error) {
	token, err := utils.GetToken()
	if err != nil {
		return nil, err
	}
	listener, err := utils.GetListener()
	if err != nil {
		return nil, err
	}
	logLevel := utils.GetExtensionLogLevel()
	var logzioLogger *logzio.LogzioSender
	if logLevel == utils.LogLevelDebug {
		logzioLogger, err = logzio.New(
			token,
			logzio.SetUrl(listener),
			logzio.SetInMemoryQueue(true),
			logzio.SetDebug(os.Stdout),
			logzio.SetinMemoryCapacity(maxBulkSizeBytes), //bytes
			logzio.SetDrainDuration(time.Second*5),
			logzio.SetDebug(os.Stdout),
		)
	} else {
		logzioLogger, err = logzio.New(
			token,
			logzio.SetUrl(listener),
			logzio.SetInMemoryQueue(true),
			logzio.SetDebug(os.Stdout),
			logzio.SetinMemoryCapacity(maxBulkSizeBytes), //bytes
			logzio.SetDrainDuration(time.Second*5),
		)
	}

	if err != nil {
		return nil, err
	}

	return logzioLogger, nil
}
