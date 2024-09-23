// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-collections/go-datastructures/queue"
	log "github.com/sirupsen/logrus"
	"logzio-lambda-extensions-logs/agent"
	"logzio-lambda-extensions-logs/extension"
	"logzio-lambda-extensions-logs/utils"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
)

// INITIAL_QUEUE_SIZE is the initial size set for the synchronous logQueue
const INITIAL_QUEUE_SIZE = 5

func main() {
	extensionName := path.Base(os.Args[0])
	printPrefix := fmt.Sprintf("[%s]", extensionName)
	setLogLevel()
	logger := log.WithFields(log.Fields{"agent": extensionName})
	logger.Debug("LogzioLogsExtension started running")
	extensionClient := extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))

	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		logger.Debug(printPrefix, "Received", s)
		logger.Info(printPrefix, "Exiting")
		cancel()
	}()

	// Register extension as soon as possible
	_, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		panic(err)
	}

	// Create Logzio Logger
	logsApiLogger, err := agent.NewLogzioLogger()
	if err != nil {
		logger.Fatal(err)
	}

	// A synchronous queue that is used to put logs from the goroutine (producer)
	// and process the logs from main goroutine (consumer)
	logQueue := queue.New(INITIAL_QUEUE_SIZE)
	// Helper function to empty the log queue
	var logsStr string = ""
	go func() {
		for {
			if !logQueue.Empty() {
				logs, err := logQueue.Get(1)
				if err != nil {
					logger.Error(printPrefix, err)
					return
				}

				logsStr = fmt.Sprintf("%v", logs[0])
				batch, err := processBatch(logsStr, logger)
				if err != nil {
					logger.Error(printPrefix, err)
					return
				}

				logger.Debugf("About to write to sender: %s", batch)
				_, err = logsApiLogger.Write([]byte(batch))
				logsApiLogger.Drain()
				if err != nil {
					logger.Error(printPrefix, err)
					return
				}
			}
		}
	}()

	// Create Logs API agent
	logsApiAgent, err := agent.NewHttpAgent(logsApiLogger, logQueue)
	if err != nil {
		logger.Fatal(err)
	}

	// Subscribe to logs API
	// Logs start being delivered only after the subscription happens.
	agentID := extensionClient.ExtensionID
	err = logsApiAgent.Init(agentID)
	if err != nil {
		logger.Fatal(err)
	}

	eventChannel := make(chan extension.EventType)
	go func() {
		for {
			logger.Info(printPrefix, " Waiting for event...")
			// This is a blocking call
			res, err := extensionClient.NextEvent(ctx)
			logger.Debugf("RECEIVED EVENT: %v", res.EventType)
			if err != nil {
				logger.Info(printPrefix, "Error:", err)
				logger.Info(printPrefix, "Exiting")
				panic(err)
			}
			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				eventChannel <- extension.Shutdown
				break
			}
		}
	}()
	// Will block until invoke or shutdown event is received or cancelled via the context.
	for {
		select {
		case <-ctx.Done():
			logger.Info(printPrefix, "Received context done event")
			logsApiLogger.Drain()
			logsApiAgent.Shutdown()
			logger.Info(printPrefix, "Exiting")
			return
		case <-eventChannel:
			logger.Info(printPrefix, "Received SHUTDOWN event")
			logsApiLogger.Drain()
			logsApiAgent.Shutdown()
			logger.Info(printPrefix, "Exiting")
			return
		default:
			continue
		}
	}
}

func processBatch(batchStr string, logger *log.Entry) (string, error) {
	logger.Debug("Processing batch")
	var batch []map[string]interface{}
	err := json.Unmarshal([]byte(batchStr), &batch)
	if err != nil {
		return "", err
	}
	logger.Debugf("Batch contains %d logs", len(batch))
	var outputBuilder strings.Builder
	for index, log := range batch {
		logzioLog := utils.ConvertLambdaLogToLogzioLog(log)
		if index > 0 {
			outputBuilder.WriteString("\n")
		}
		bytes, err := json.Marshal(logzioLog)
		if err != nil {
			return "", err
		}
		outputBuilder.Write(bytes)
	}

	return outputBuilder.String(), nil
}

func setLogLevel() {
	logLevelStr := utils.GetExtensionLogLevel()
	logLevel, err := log.ParseLevel(logLevelStr)
	if err == nil {
		// if no error occurred while trying to parse the log level, we'll set the user's chosen log level.
		// otherwise, we'll use the logger's default log level (info)
		log.SetLevel(logLevel)
	}
}
