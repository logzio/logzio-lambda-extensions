package utils

import (
	"encoding/json"
	"fmt"
	"github.com/vjeantet/grok"
)

const (
	fldLogzioTimestamp    = "@timestamp"
	fldLambdaTime         = "time"
	fldLogzioType         = "type"
	fldLambdaType         = "type"
	fldLogzioLambdaType   = "lambda.log.type"
	fldLambdaRecord       = "record"
	fldLogzioMsg          = "message"
	fldLogzioLambdaRecord = "lambda.record"

	extensionType = "lambda-extension-logs"

	grokKeyLogFormat = "LOG_PATTERN"
)

// ConvertLambdaLogToLogzioLog converts a log that was sent from AWS Logs API to a log in a Logz.io format
func ConvertLambdaLogToLogzioLog(lambdaLog map[string]interface{}) map[string]interface{} {
	sendAsString := false
	logzioLog := make(map[string]interface{})
	logzioLog[fldLogzioTimestamp] = lambdaLog[fldLambdaTime]
	logzioLog[fldLogzioType] = extensionType
	logzioLog[fldLogzioLambdaType] = lambdaLog[fldLambdaType]
	logger.Debugf("working on: %v", lambdaLog[fldLambdaRecord])

	switch lambdaLog[fldLambdaRecord].(type) {
	case string:
		grokPattern := GetGrokPatterns()
		logsFormat := GetLogsFormat()
		if len(grokPattern) > 0 && len(logsFormat) > 0 {
			logger.Debugf("grok pattern: %s", grokPattern)
			logger.Debugf("logs format: %s", logsFormat)
			logger.Info("detected grok pattern and logs format. trying to parse log")
			err := parseFields(logzioLog, lambdaLog[fldLambdaRecord].(string), grokPattern, logsFormat)
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
			logzioLog[fldLogzioMsg] = lambdaLog[fldLambdaRecord]
		}
	default:
		logzioLog[fldLogzioLambdaRecord] = lambdaLog[fldLambdaRecord]
	}

	return logzioLog
}

func parseFields(logMap map[string]interface{}, fieldsToParse, grokPatterns, logsFormat string) error {
	g, err := grok.New()
	if err != nil {
		return err
	}

	err = addGrokPatterns(g, grokPatterns, logsFormat)
	if err != nil {
		return err
	}

	logger.Debugf("about to parse: %s", fieldsToParse)
	fields, err := g.Parse(fmt.Sprintf("%%{%s}", grokKeyLogFormat), fieldsToParse)
	logger.Debugf("number of fields after grok: %d", len(fields))
	if err != nil {
		return err
	}

	if len(fields) == 0 {
		return fmt.Errorf("could not parse fields with the current patterns & format")
	}

	for key, val := range fields {
		if key != grokKeyLogFormat {
			logger.Debugf("adding field: %s to logzio log", key)
			logMap[fldLogzioLambdaRecord+"."+key] = val
		} else {
			logger.Debugf("ignoring field of %s", grokKeyLogFormat)
		}
	}

	return nil
}

func addGrokPatterns(g *grok.Grok, patternsStr, logFormat string) error {
	var grokPatterns map[string]string
	err := json.Unmarshal([]byte(patternsStr), &grokPatterns)
	if err != nil {
		return err
	}

	err = g.AddPatternsFromMap(grokPatterns)
	if err != nil {
		return err
	}

	logger.Debugf("added patterns from user")

	err = g.AddPattern(grokKeyLogFormat, logFormat)
	if err != nil {
		return err
	}

	logger.Debugf("added %s: %s", grokKeyLogFormat, logFormat)

	return nil
}
