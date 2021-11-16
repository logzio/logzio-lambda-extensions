package utils

import (
	"encoding/json"
	"fmt"
	"github.com/vjeantet/grok"
)

const (
	FldLogzioTimestamp = "@timestamp"
	FldLambdaTime      = "time"
	FldLogzioType       = "type"
	FldLambdaType       = "type"
	FldLogzioLambdaType = "lambda.log.type"
	FldLambdaRecord    = "record"
	FldLogzioMsg          = "message"
	FldLogzioMsgNested    = "message_nested"
	FldLogzioLambdaRecord = "lambda.record"

	ExtensionType = "lambda-extension-logs"

	grokKeyLogFormat = "LOG_FORMAT"
)

// ConvertLambdaLogToLogzioLog converts a log that was sent from AWS Logs API to a log in a Logz.io format
func ConvertLambdaLogToLogzioLog(lambdaLog map[string]interface{}) map[string]interface{} {
	sendAsString := false
	logzioLog := make(map[string]interface{})
	logzioLog[FldLogzioTimestamp] = lambdaLog[FldLambdaTime]
	logzioLog[FldLogzioType] = ExtensionType
	logzioLog[FldLogzioLambdaType] = lambdaLog[FldLambdaType]
	logger.Debugf("working on: %v", lambdaLog[FldLambdaRecord])

	switch lambdaLog[FldLambdaRecord].(type) {
	case string:
		grokPattern := GetGrokPatterns()
		logsFormat := GetLogsFormat()
		if len(grokPattern) > 0 && len(logsFormat) > 0 {
			logger.Debugf("grok pattern: %s", grokPattern)
			logger.Debugf("logs format: %s", logsFormat)
			logger.Info("detected grok pattern and logs format. trying to parse log")
			err := parseFields(logzioLog, lambdaLog[FldLambdaRecord].(string), grokPattern, logsFormat)
			if err != nil {
				logger.Errorf("error occurred while trying to parse fields. sedning log as a string: %s", err.Error())
				sendAsString = true
			}
		} else {
			if len(grokPattern) > 0 || len(logsFormat) > 0 {
				logger.Error("grok pattern and logs format must be set in order to parse fields. sending log as string.")
			}

			sendAsString = true
		}

		if sendAsString {
			var nested map[string]interface{}
			err := json.Unmarshal([]byte(fmt.Sprintf(`%s`, lambdaLog[FldLambdaRecord])), &nested)
			if err != nil {
				logger.Infof("error occurred while checking if log %s is JSON. ignore if this is not JSON: %s", lambdaLog[FldLambdaRecord], err.Error())
				logzioLog[FldLogzioMsg] = lambdaLog[FldLambdaRecord]
			} else {
				logger.Debugf("detected JSON: %s", lambdaLog[FldLambdaRecord])
				logzioLog[FldLogzioMsgNested] = nested
			}
		}
	default:
		logzioLog[FldLogzioLambdaRecord] = lambdaLog[FldLambdaRecord]
	}

	return logzioLog
}

func parseFields(logMap map[string]interface{}, fieldsToParse, grokPatterns, logsFormat string) error {
	g, err := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		return err
	}

	err = addGrokPatterns(g, grokPatterns, logsFormat)
	if err != nil {
		return err
	}

	logger.Debugf("about to parse: %s", fieldsToParse)
	fields, err := g.Parse(fmt.Sprintf(`%%{%s}`, grokKeyLogFormat), fmt.Sprintf(`%s`, fieldsToParse))
	logger.Debugf("number of fields after grok: %d", len(fields))
	if err != nil {
		return err
	}

	if len(fields) == 0 {
		return fmt.Errorf("could not parse fields with the current patterns & format")
	}

	addFields(logMap, fields)

	return nil
}

func addGrokPatterns(g *grok.Grok, patternsStr, logFormat string) error {
	var grokPatterns map[string]string
	err := json.Unmarshal([]byte(patternsStr), &grokPatterns)
	if err != nil {
		return err
	}

	for key, val := range grokPatterns {
		fVal := fmt.Sprintf(`%s`, val)
		logger.Debugf("adding pattern %s", fVal)
		g.AddPattern(key, fVal)
	}

	logger.Debugf("added patterns from user")

	err = g.AddPattern(grokKeyLogFormat, fmt.Sprintf(`%s`, logFormat))
	if err != nil {
		return err
	}

	logger.Debugf("added %s: %s", grokKeyLogFormat, logFormat)

	return nil
}

func addFields(logsMap map[string]interface{}, fields map[string]string) {
	var nested map[string]interface{}
	for key, val := range fields {
		logger.Debugf("adding field: %s to logzio log", key)
		// Trying to see if the string is in JSON format.
		// If so - add the nested version to the log
		err := json.Unmarshal([]byte(fmt.Sprintf(`%s`, val)), &nested)
		if err != nil {
			logger.Infof("error occurred while checking if log %s is JSON. ignore if this is not JSON: %s", val, err.Error())
		} else {
			logger.Debugf("detected JSON: %s", val)
		}

		if nested != nil && len(nested) > 0 {
			logsMap[key] = nested
		} else {
			logsMap[key] = val
		}
	}
}
