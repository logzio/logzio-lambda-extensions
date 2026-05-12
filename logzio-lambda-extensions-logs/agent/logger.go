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
	opts := []logzio.SenderOptionFunc{
		logzio.SetUrl(listener),
		logzio.SetInMemoryQueue(true),
		logzio.SetinMemoryCapacity(maxBulkSizeBytes),
		logzio.SetDrainDuration(time.Second * 5),
	}
	if utils.GetExtensionLogLevel() == utils.LogLevelDebug {
		opts = append(opts, logzio.SetDebug(os.Stdout))
	}
	logzioLogger, err := logzio.New(token, opts...)

	if err != nil {
		return nil, err
	}

	return logzioLogger, nil
}
